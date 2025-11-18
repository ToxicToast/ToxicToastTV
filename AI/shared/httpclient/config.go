package httpclient

import "time"

// Config holds the configuration for the HTTP client
type Config struct {
	// Timeout is the maximum time a request can take
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryWaitMin is the minimum wait time between retries
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum wait time between retries
	RetryWaitMax time.Duration

	// Headers are default headers to send with every request
	Headers map[string]string

	// UserAgent is the User-Agent header value
	UserAgent string
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
		Headers:      make(map[string]string),
		UserAgent:    "ToxicToastGo-HTTPClient/1.0",
	}
}
