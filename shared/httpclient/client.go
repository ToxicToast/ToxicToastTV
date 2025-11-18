package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client is an HTTP client with retry capabilities
type Client struct {
	httpClient      *http.Client
	config          *Config
	backoffStrategy BackoffStrategy
}

// New creates a new HTTP client with the given config
func New(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config:          config,
		backoffStrategy: ExponentialBackoff,
	}
}

// WithBackoffStrategy sets a custom backoff strategy
func (c *Client) WithBackoffStrategy(strategy BackoffStrategy) *Client {
	c.backoffStrategy = strategy
	return c
}

// Get performs a GET request with retries
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return c.Do(ctx, "GET", url, nil, headers)
}

// Post performs a POST request with retries
func (c *Client) Post(ctx context.Context, url string, body []byte, headers map[string]string) (*http.Response, error) {
	return c.Do(ctx, "POST", url, body, headers)
}

// Put performs a PUT request with retries
func (c *Client) Put(ctx context.Context, url string, body []byte, headers map[string]string) (*http.Response, error) {
	return c.Do(ctx, "PUT", url, body, headers)
}

// Delete performs a DELETE request with retries
func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return c.Do(ctx, "DELETE", url, nil, headers)
}

// Do performs an HTTP request with retries
func (c *Client) Do(ctx context.Context, method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Create request
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set default headers
		req.Header.Set("User-Agent", c.config.UserAgent)
		for k, v := range c.config.Headers {
			req.Header.Set(k, v)
		}

		// Set request-specific headers (override defaults)
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// Execute request
		startTime := time.Now()
		resp, err = c.httpClient.Do(req)
		duration := time.Since(startTime)

		// Log request
		log.Printf("HTTP %s %s - Status: %v - Duration: %v - Attempt: %d/%d",
			method, url, getStatusCode(resp), duration, attempt+1, c.config.MaxRetries+1)

		// Check for errors
		if err != nil {
			lastErr = err
			if attempt < c.config.MaxRetries {
				waitDuration := c.backoffStrategy(attempt, c.config.RetryWaitMin, c.config.RetryWaitMax)
				log.Printf("Request failed (attempt %d/%d): %v - Retrying in %v",
					attempt+1, c.config.MaxRetries+1, err, waitDuration)
				time.Sleep(waitDuration)
				continue
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", attempt+1, err)
		}

		// Check HTTP status code
		if resp.StatusCode >= 400 {
			// Read body for error message
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			httpErr := &HTTPError{
				StatusCode: resp.StatusCode,
				Body:       bodyBytes,
				URL:        url,
				Method:     method,
			}

			// Check if retryable
			if IsRetryable(resp.StatusCode) && attempt < c.config.MaxRetries {
				lastErr = httpErr
				waitDuration := c.backoffStrategy(attempt, c.config.RetryWaitMin, c.config.RetryWaitMax)
				log.Printf("HTTP %d error (attempt %d/%d) - Retrying in %v",
					resp.StatusCode, attempt+1, c.config.MaxRetries+1, waitDuration)
				time.Sleep(waitDuration)
				continue
			}

			return nil, httpErr
		}

		// Success
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// GetJSON is a convenience method for GET requests that expect JSON
func (c *Client) GetJSON(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Accept"] = "application/json"

	resp, err := c.Get(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// PostJSON is a convenience method for POST requests with JSON
func (c *Client) PostJSON(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	resp, err := c.Post(ctx, url, body, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, nil
}

// Helper function to get status code from response
func getStatusCode(resp *http.Response) string {
	if resp == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", resp.StatusCode)
}
