package httpclient

import (
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy determines the wait time between retries
type BackoffStrategy func(attempt int, min, max time.Duration) time.Duration

// ExponentialBackoff returns an exponential backoff duration with jitter
func ExponentialBackoff(attempt int, min, max time.Duration) time.Duration {
	// Calculate exponential backoff: min * 2^attempt
	backoff := float64(min) * math.Pow(2, float64(attempt))

	// Add jitter (random value between 0 and backoff)
	jitter := rand.Float64() * backoff

	// Calculate total wait time
	wait := time.Duration(backoff + jitter)

	// Cap at max
	if wait > max {
		wait = max
	}

	return wait
}

// LinearBackoff returns a linear backoff duration
func LinearBackoff(attempt int, min, max time.Duration) time.Duration {
	wait := min * time.Duration(attempt+1)
	if wait > max {
		wait = max
	}
	return wait
}

// ConstantBackoff returns a constant backoff duration
func ConstantBackoff(attempt int, min, max time.Duration) time.Duration {
	return min
}
