package httpclient

import "fmt"

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Body       []byte
	URL        string
	Method     string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s %s - %s", e.StatusCode, e.Method, e.URL, string(e.Body))
}

// IsRetryable checks if an HTTP status code is retryable
func IsRetryable(statusCode int) bool {
	// Retry on:
	// - 429 Too Many Requests
	// - 500 Internal Server Error
	// - 502 Bad Gateway
	// - 503 Service Unavailable
	// - 504 Gateway Timeout
	switch statusCode {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}
