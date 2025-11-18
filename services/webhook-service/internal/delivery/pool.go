package delivery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"

	"github.com/toxictoast/toxictoastgo/shared/logger"
)

// Pool manages a pool of delivery workers
type Pool struct {
	worker           *Worker
	deliveryRepo     interfaces.DeliveryRepository
	webhookRepo      interfaces.WebhookRepository
	workerCount      int
	deliveryQueue    chan *QueuedDelivery
	retryQueue       chan *QueuedDelivery
	wg               sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
	retryCheckTicker *time.Ticker
}

type QueuedDelivery struct {
	Delivery *domain.Delivery
	Webhook  *domain.Webhook
}

type PoolConfig struct {
	WorkerCount      int
	QueueSize        int
	RetryCheckInterval time.Duration
}

func NewPool(
	worker *Worker,
	deliveryRepo interfaces.DeliveryRepository,
	webhookRepo interfaces.WebhookRepository,
	config PoolConfig,
) *Pool {
	if config.WorkerCount == 0 {
		config.WorkerCount = 10
	}
	if config.QueueSize == 0 {
		config.QueueSize = 1000
	}
	if config.RetryCheckInterval == 0 {
		config.RetryCheckInterval = 1 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		worker:           worker,
		deliveryRepo:     deliveryRepo,
		webhookRepo:      webhookRepo,
		workerCount:      config.WorkerCount,
		deliveryQueue:    make(chan *QueuedDelivery, config.QueueSize),
		retryQueue:       make(chan *QueuedDelivery, config.QueueSize),
		ctx:              ctx,
		cancel:           cancel,
		retryCheckTicker: time.NewTicker(config.RetryCheckInterval),
	}
}

// Start starts the worker pool
func (p *Pool) Start() {
	logger.Info(fmt.Sprintf("Starting delivery pool with %d workers", p.workerCount))

	// Start workers for new deliveries
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.deliveryWorker(i)
	}

	// Start workers for retries (use half the workers)
	retryWorkers := p.workerCount / 2
	if retryWorkers < 1 {
		retryWorkers = 1
	}
	for i := 0; i < retryWorkers; i++ {
		p.wg.Add(1)
		go p.retryWorker(i)
	}

	// Start retry checker
	p.wg.Add(1)
	go p.retryChecker()

	logger.Info("Delivery pool started successfully")
}

// Stop gracefully stops the worker pool
func (p *Pool) Stop() {
	logger.Info("Stopping delivery pool...")
	p.retryCheckTicker.Stop()
	p.cancel()
	close(p.deliveryQueue)
	close(p.retryQueue)
	p.wg.Wait()
	logger.Info("Delivery pool stopped")
}

// QueueDelivery queues a delivery for processing
func (p *Pool) QueueDelivery(delivery *domain.Delivery, webhook *domain.Webhook) error {
	select {
	case p.deliveryQueue <- &QueuedDelivery{Delivery: delivery, Webhook: webhook}:
		logger.Info(fmt.Sprintf("Queued delivery %s for webhook %s", delivery.ID, webhook.ID))
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("pool is shutting down")
	default:
		return fmt.Errorf("delivery queue is full")
	}
}

// deliveryWorker processes deliveries from the queue
func (p *Pool) deliveryWorker(id int) {
	defer p.wg.Done()
	logger.Info(fmt.Sprintf("Delivery worker %d started", id))

	for {
		select {
		case <-p.ctx.Done():
			logger.Info(fmt.Sprintf("Delivery worker %d stopped", id))
			return
		case queued, ok := <-p.deliveryQueue:
			if !ok {
				logger.Info(fmt.Sprintf("Delivery worker %d stopped (queue closed)", id))
				return
			}

			if queued == nil {
				continue
			}

			logger.Info(fmt.Sprintf("Worker %d processing delivery %s", id, queued.Delivery.ID))

			// Process delivery
			err := p.worker.DeliverWebhook(p.ctx, queued.Delivery, queued.Webhook)

			// Update webhook statistics
			success := err == nil
			if statsErr := p.webhookRepo.UpdateStatistics(p.ctx, queued.Webhook.ID, success); statsErr != nil {
				logger.Error(fmt.Sprintf("Failed to update webhook statistics: %v", statsErr))
			}

			if err != nil {
				logger.Error(fmt.Sprintf("Worker %d failed to deliver webhook %s: %v", id, queued.Delivery.ID, err))
			} else {
				logger.Info(fmt.Sprintf("Worker %d successfully delivered webhook %s", id, queued.Delivery.ID))
			}
		}
	}
}

// retryWorker processes retry deliveries
func (p *Pool) retryWorker(id int) {
	defer p.wg.Done()
	logger.Info(fmt.Sprintf("Retry worker %d started", id))

	for {
		select {
		case <-p.ctx.Done():
			logger.Info(fmt.Sprintf("Retry worker %d stopped", id))
			return
		case queued, ok := <-p.retryQueue:
			if !ok {
				logger.Info(fmt.Sprintf("Retry worker %d stopped (queue closed)", id))
				return
			}

			if queued == nil {
				continue
			}

			logger.Info(fmt.Sprintf("Retry worker %d processing delivery %s", id, queued.Delivery.ID))

			// Process retry
			err := p.worker.ProcessRetry(p.ctx, queued.Delivery)

			// Update webhook statistics
			success := err == nil
			if statsErr := p.webhookRepo.UpdateStatistics(p.ctx, queued.Webhook.ID, success); statsErr != nil {
				logger.Error(fmt.Sprintf("Failed to update webhook statistics: %v", statsErr))
			}

			if err != nil {
				logger.Error(fmt.Sprintf("Retry worker %d failed to deliver webhook %s: %v", id, queued.Delivery.ID, err))
			} else {
				logger.Info(fmt.Sprintf("Retry worker %d successfully delivered webhook %s", id, queued.Delivery.ID))
			}
		}
	}
}

// retryChecker periodically checks for deliveries that need retry
func (p *Pool) retryChecker() {
	defer p.wg.Done()
	logger.Info("Retry checker started")

	for {
		select {
		case <-p.ctx.Done():
			logger.Info("Retry checker stopped")
			return
		case <-p.retryCheckTicker.C:
			p.checkPendingRetries()
		}
	}
}

// checkPendingRetries fetches and queues deliveries that need retry
func (p *Pool) checkPendingRetries() {
	ctx := context.Background()

	// Fetch pending retries (limit to 100 at a time)
	deliveries, err := p.deliveryRepo.GetPendingRetries(ctx, 100)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to fetch pending retries: %v", err))
		return
	}

	if len(deliveries) == 0 {
		logger.Info("No pending retries found")
		return
	}

	logger.Info(fmt.Sprintf("Found %d deliveries pending retry", len(deliveries)))

	// Queue each delivery for retry
	for _, delivery := range deliveries {
		// Ensure webhook is loaded
		if delivery.Webhook == nil {
			webhook, err := p.webhookRepo.GetByID(ctx, delivery.WebhookID)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to fetch webhook %s: %v", delivery.WebhookID, err))
				continue
			}
			delivery.Webhook = webhook
		}

		// Queue for retry
		select {
		case p.retryQueue <- &QueuedDelivery{Delivery: delivery, Webhook: delivery.Webhook}:
			logger.Info(fmt.Sprintf("Queued delivery %s for retry", delivery.ID))
		case <-p.ctx.Done():
			return
		default:
			logger.Info(fmt.Sprintf("Retry queue full, skipping delivery %s", delivery.ID))
		}
	}
}

// GetQueueStatus returns the current queue sizes
func (p *Pool) GetQueueStatus() (deliveryQueueSize, retryQueueSize int) {
	return len(p.deliveryQueue), len(p.retryQueue)
}
