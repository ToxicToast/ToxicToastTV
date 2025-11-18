# HTTP Client

A robust HTTP client with automatic retry logic, exponential backoff, and configurable timeouts.

## Features

- ✅ Automatic retries with exponential backoff
- ✅ Configurable timeouts
- ✅ Custom headers support
- ✅ JSON convenience methods
- ✅ Context support for cancellation
- ✅ Detailed logging
- ✅ Multiple backoff strategies

## Usage

### Basic Usage

```go
import "github.com/toxictoast/toxictoastgo/shared/httpclient"

// Create client with default config
client := httpclient.New(nil)

// Make a GET request
resp, err := client.Get(context.Background(), "https://api.example.com/data", nil)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
```

### Custom Configuration

```go
config := &httpclient.Config{
    Timeout:      10 * time.Second,
    MaxRetries:   5,
    RetryWaitMin: 2 * time.Second,
    RetryWaitMax: 60 * time.Second,
    UserAgent:    "MyApp/1.0",
    Headers: map[string]string{
        "X-API-Key": "your-api-key",
    },
}

client := httpclient.New(config)
```

### JSON Requests

```go
// GET JSON
data, err := client.GetJSON(ctx, "https://api.example.com/users", nil)
if err != nil {
    log.Fatal(err)
}

// POST JSON
body := []byte(`{"name": "John"}`)
respData, err := client.PostJSON(ctx, "https://api.example.com/users", body, nil)
```

### Custom Backoff Strategy

```go
client := httpclient.New(config).
    WithBackoffStrategy(httpclient.LinearBackoff)
```

## Retryable Status Codes

The client automatically retries on these HTTP status codes:
- 429 Too Many Requests
- 500 Internal Server Error
- 502 Bad Gateway
- 503 Service Unavailable
- 504 Gateway Timeout

## Available Backoff Strategies

- `ExponentialBackoff` (default) - Exponential backoff with jitter
- `LinearBackoff` - Linear increase in wait time
- `ConstantBackoff` - Fixed wait time between retries
