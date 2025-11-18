package delivery

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"

	"github.com/toxictoast/toxictoastgo/shared/logger"
)

// Worker handles webhook deliveries with retry logic
type Worker struct {
	httpClient       *http.Client
	deliveryRepo     interfaces.DeliveryRepository
	maxRetries       int
	initialRetryDelay time.Duration
	maxRetryDelay    time.Duration
	timeout          time.Duration
}

type WorkerConfig struct {
	MaxRetries       int
	InitialRetryDelay time.Duration
	MaxRetryDelay    time.Duration
	Timeout          time.Duration
}

func NewWorker(
	deliveryRepo interfaces.DeliveryRepository,
	config WorkerConfig,
) *Worker {
	if config.MaxRetries == 0 {
		config.MaxRetries = 5
	}
	if config.InitialRetryDelay == 0 {
		config.InitialRetryDelay = 5 * time.Second
	}
	if config.MaxRetryDelay == 0 {
		config.MaxRetryDelay = 5 * time.Minute
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Worker{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		deliveryRepo:      deliveryRepo,
		maxRetries:        config.MaxRetries,
		initialRetryDelay: config.InitialRetryDelay,
		maxRetryDelay:     config.MaxRetryDelay,
		timeout:           config.Timeout,
	}
}

// DeliverWebhook attempts to deliver a webhook with retry logic
func (w *Worker) DeliverWebhook(ctx context.Context, delivery *domain.Delivery, webhook *domain.Webhook) error {
	logger.Info(fmt.Sprintf("Starting delivery %s to webhook %s (%s)", delivery.ID, webhook.ID, webhook.URL))

	// Attempt delivery with retries
	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		success, responseStatus, responseBody, duration, err := w.attemptDelivery(ctx, delivery, webhook, attempt)

		// Record attempt
		attemptRecord := &domain.DeliveryAttempt{
			DeliveryID:     delivery.ID,
			AttemptNumber:  attempt,
			RequestURL:     webhook.URL,
			ResponseStatus: responseStatus,
			ResponseBody:   responseBody,
			Success:        success,
			Error:          "",
			DurationMs:     int(duration.Milliseconds()),
		}

		if err != nil {
			attemptRecord.Error = err.Error()
		}

		if err := w.deliveryRepo.CreateAttempt(ctx, attemptRecord); err != nil {
			logger.Error(fmt.Sprintf("Failed to record delivery attempt: %v", err))
		}

		// Update delivery status
		delivery.AttemptCount = attempt
		now := time.Now()
		delivery.LastAttemptAt = &now

		if success {
			// Success - mark as delivered
			delivery.Status = domain.DeliveryStatusSuccess
			delivery.NextRetryAt = nil

			if err := w.deliveryRepo.Update(ctx, delivery); err != nil {
				logger.Error(fmt.Sprintf("Failed to update delivery status: %v", err))
			}

			logger.Info(fmt.Sprintf("Successfully delivered webhook %s (attempt %d/%d)", delivery.ID, attempt, w.maxRetries))
			return nil
		}

		// Failed - check if we should retry
		if attempt < w.maxRetries {
			// Calculate next retry time with exponential backoff
			retryDelay := w.calculateRetryDelay(attempt)
			nextRetry := time.Now().Add(retryDelay)
			delivery.NextRetryAt = &nextRetry
			delivery.Status = domain.DeliveryStatusRetrying
			delivery.LastError = fmt.Sprintf("Attempt %d failed: %v", attempt, err)

			if err := w.deliveryRepo.Update(ctx, delivery); err != nil {
				logger.Error(fmt.Sprintf("Failed to update delivery status: %v", err))
			}

			logger.Info(fmt.Sprintf("Delivery %s failed (attempt %d/%d), will retry in %v",
				delivery.ID, attempt, w.maxRetries, retryDelay))

			// Wait before next attempt (unless context is cancelled)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continue to next attempt
			}
		} else {
			// Max retries reached - mark as failed
			delivery.Status = domain.DeliveryStatusFailed
			delivery.NextRetryAt = nil
			delivery.LastError = fmt.Sprintf("Max retries reached. Last error: %v", err)

			if err := w.deliveryRepo.Update(ctx, delivery); err != nil {
				logger.Error(fmt.Sprintf("Failed to update delivery status: %v", err))
			}

			logger.Error(fmt.Sprintf("Failed to deliver webhook %s after %d attempts", delivery.ID, w.maxRetries))
			return fmt.Errorf("max retries reached: %w", err)
		}
	}

	return nil
}

// attemptDelivery makes a single HTTP POST attempt
func (w *Worker) attemptDelivery(
	ctx context.Context,
	delivery *domain.Delivery,
	webhook *domain.Webhook,
	attemptNumber int,
) (success bool, responseStatus int, responseBody string, duration time.Duration, err error) {
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBufferString(delivery.EventPayload))
	if err != nil {
		return false, 0, "", time.Since(startTime), fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ToxicToastGo-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", delivery.EventType)
	req.Header.Set("X-Webhook-Delivery", delivery.ID)
	req.Header.Set("X-Webhook-Attempt", fmt.Sprintf("%d", attemptNumber))
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Generate and add HMAC signature
	signature := GenerateSignature([]byte(delivery.EventPayload), webhook.Secret)
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Signature-256", "sha256="+signature) // GitHub-style format

	// Make request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return false, 0, "", time.Since(startTime), fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (limit to 10KB to prevent memory issues)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to read response body: %v", err))
		bodyBytes = []byte{}
	}
	responseBody = string(bodyBytes)

	duration = time.Since(startTime)
	responseStatus = resp.StatusCode

	// Consider 2xx status codes as success
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, responseStatus, responseBody, duration, nil
	}

	return false, responseStatus, responseBody, duration, fmt.Errorf("HTTP %d: %s", resp.StatusCode, responseBody)
}

// calculateRetryDelay calculates exponential backoff delay
func (w *Worker) calculateRetryDelay(attemptNumber int) time.Duration {
	// Exponential backoff: initialDelay * 2^(attempt-1)
	delay := float64(w.initialRetryDelay) * math.Pow(2, float64(attemptNumber-1))
	retryDelay := time.Duration(delay)

	// Cap at max delay
	if retryDelay > w.maxRetryDelay {
		retryDelay = w.maxRetryDelay
	}

	return retryDelay
}

// ProcessRetry processes a single retry delivery
func (w *Worker) ProcessRetry(ctx context.Context, delivery *domain.Delivery) error {
	// Fetch the webhook
	webhook := delivery.Webhook
	if webhook == nil {
		return fmt.Errorf("webhook not loaded for delivery %s", delivery.ID)
	}

	// Check if webhook is still active
	if !webhook.Active {
		delivery.Status = domain.DeliveryStatusFailed
		delivery.LastError = "Webhook is no longer active"
		if err := w.deliveryRepo.Update(ctx, delivery); err != nil {
			logger.Error(fmt.Sprintf("Failed to update delivery: %v", err))
		}
		return fmt.Errorf("webhook %s is not active", webhook.ID)
	}

	// Continue delivery from current attempt count
	return w.DeliverWebhook(ctx, delivery, webhook)
}
