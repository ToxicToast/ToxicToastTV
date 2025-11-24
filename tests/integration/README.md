# Integration Tests

This directory contains end-to-end integration tests for the ToxicToastGo microservices ecosystem.

## Overview

The integration tests verify the complete functionality of the system by testing:
- **Auth Flow E2E**: Complete authentication flow from registration to logout
- **Kafka Events**: Event publishing and consumption across services
- **Service Communication**: gRPC and HTTP communication between services

## Prerequisites

Before running integration tests, ensure all services are running:

1. **PostgreSQL Database** (port 5432)
   - user-service database
   - auth-service database

2. **Kafka/Redpanda** (port 19092)

3. **Microservices**:
   - user-service (gRPC port 50051)
   - auth-service (gRPC port 50052)
   - gateway-service (HTTP port 8080)

## Quick Start

### 1. Start Infrastructure

```bash
# Using Docker Compose (recommended)
cd ToxicToastGo
docker-compose up -d postgres kafka

# Wait for services to be ready
sleep 10
```

### 2. Start Services

```bash
# Terminal 1: User Service
cd services/user-service
go run cmd/server/main.go

# Terminal 2: Auth Service
cd services/auth-service
go run cmd/server/main.go

# Terminal 3: Gateway Service
cd services/gateway-service
go run cmd/server/main.go
```

### 3. Run Integration Tests

```bash
cd tests/integration

# Download dependencies
go mod download

# Run all integration tests
go test -v

# Run specific test
go test -v -run TestAuthFlowE2E

# Run only short tests (skips integration tests)
go test -v -short
```

## Test Suites

### Auth Flow E2E Test (`auth_flow_test.go`)

Tests the complete authentication flow:

1. ✅ **Register** - Create new user account
2. ✅ **Login** - Authenticate with credentials
3. ✅ **Validate Token** - Verify JWT token validity
4. ✅ **Access Protected Endpoint** - Use token to access `/auth/me`
5. ✅ **Access Public Endpoint** - Test unauthenticated access
6. ✅ **Access Protected Without Token** - Verify authentication required
7. ✅ **Access Protected With Token** - Verify token grants access
8. ✅ **Refresh Token** - Request new access token
9. ✅ **Access With New Token** - Verify refreshed token works
10. ✅ **Logout** - Revoke token
11. ✅ **Access After Logout** - Verify token is revoked

**Run:**
```bash
go test -v -run TestAuthFlowE2E
```

**Expected Duration:** ~5-10 seconds

### Kafka Events Test (`kafka_events_test.go`)

Tests Kafka event publishing and consumption:

#### TestKafkaUserEvents
- Verifies `user.created` event is published
- Verifies `auth.registered` event is published
- Verifies `auth.login` event is published
- Validates event structure and fields

#### TestKafkaEventStructure
- Validates event schema for all topics
- Ensures required fields are present

#### TestKafkaTopicsExist
- Verifies all expected Kafka topics exist
- Lists available topics

**Run:**
```bash
go test -v -run TestKafka
```

**Expected Duration:** ~30-60 seconds

## Configuration

### Service URLs

Default configuration (adjust if needed):

```go
const (
    gatewayBaseURL = "http://localhost:8080/api"
    kafkaBroker    = "localhost:19092"
    testTimeout    = 30 * time.Second
)
```

### Environment Variables

Override defaults with environment variables:

```bash
export GATEWAY_URL="http://localhost:8080/api"
export KAFKA_BROKER="localhost:19092"
export TEST_TIMEOUT="60s"
```

## Troubleshooting

### Services Not Running

**Error:** `Failed to connect to http://localhost:8080`

**Solution:** Ensure gateway-service is running on port 8080

```bash
cd services/gateway-service
go run cmd/server/main.go
```

### Kafka Connection Failed

**Error:** `Failed to connect to Kafka: dial tcp [::1]:19092: connect: connection refused`

**Solution:** Start Kafka/Redpanda

```bash
docker-compose up -d kafka
# Wait for Kafka to be ready
sleep 10
```

### Database Connection Failed

**Error:** `failed to connect to database`

**Solution:** Start PostgreSQL and create databases

```bash
docker-compose up -d postgres

# Create databases
psql -h localhost -U postgres -c "CREATE DATABASE user_service;"
psql -h localhost -U postgres -c "CREATE DATABASE auth_service;"
```

### Test User Already Exists

**Error:** `email already exists`

**Solution:** Tests use unique timestamps, but if running very quickly, you may hit conflicts. Wait 1 second and retry.

### Events Not Found in Kafka

**Warning:** `⚠ user.created event not found (may have been consumed already)`

**Explanation:** This is normal if events were consumed by other consumers. The test reads from the latest offset.

**Solution:** No action needed - this is informational only.

## Continuous Integration

### GitHub Actions

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432

      kafka:
        image: docker.redpanda.com/redpandadata/redpanda:latest
        ports:
          - 19092:19092

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Start Services
        run: |
          cd services/user-service && go run cmd/server/main.go &
          cd services/auth-service && go run cmd/server/main.go &
          cd services/gateway-service && go run cmd/server/main.go &
          sleep 10

      - name: Run Integration Tests
        run: |
          cd tests/integration
          go test -v
```

## Test Coverage

Integration tests verify:

- ✅ User registration flow
- ✅ User authentication (login/logout)
- ✅ JWT token generation and validation
- ✅ Token refresh mechanism
- ✅ Token revocation
- ✅ Protected endpoint access control
- ✅ Public endpoint accessibility
- ✅ Kafka event publishing (user.*, auth.*)
- ✅ Event schema validation
- ✅ Service-to-service communication

## Writing New Tests

### Test Template

```go
func TestMyIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    client := &http.Client{Timeout: testTimeout}

    t.Run("SubTest", func(t *testing.T) {
        // Your test logic here
        resp, err := makeJSONRequest(client, "GET", url, nil, "")
        if err != nil {
            t.Fatalf("Request failed: %v", err)
        }

        if resp.StatusCode != http.StatusOK {
            t.Errorf("Expected 200, got %d", resp.StatusCode)
        }

        t.Logf("✓ Test passed")
    })
}
```

### Best Practices

1. **Use Unique Test Data** - Generate unique usernames/emails with timestamps
2. **Clean Up Resources** - Use `defer` to clean up connections
3. **Set Timeouts** - Always use context with timeout
4. **Check Status Codes** - Verify HTTP status codes
5. **Validate Response Structure** - Check JSON structure and required fields
6. **Log Progress** - Use `t.Logf()` for debugging
7. **Skip in Short Mode** - Use `if testing.Short() { t.Skip() }`

## Performance Benchmarks

### Auth Flow E2E

- Registration: ~200-500ms
- Login: ~150-300ms
- Token validation: ~50-100ms
- Protected endpoint: ~100-200ms
- Token refresh: ~150-300ms

### Kafka Events

- Event publish: ~10-50ms
- Event consume: ~100-500ms (depends on consumer lag)

## Monitoring

### View Test Logs

```bash
# Verbose output with timestamps
go test -v -timeout 5m 2>&1 | tee integration-test.log

# JSON output for parsing
go test -json > results.json
```

### Test Metrics

```bash
# Test coverage
go test -cover

# Test benchmarks
go test -bench=. -benchmem
```

## Support

For issues or questions:
- Check service logs
- Verify all services are running
- Check database connectivity
- Verify Kafka topics exist

## Future Enhancements

- [ ] Add RBAC integration tests
- [ ] Add user update/delete flow tests
- [ ] Add concurrent user tests (load testing)
- [ ] Add service resilience tests (circuit breaker, retry)
- [ ] Add integration with other services (blog, foodfolio, etc.)
