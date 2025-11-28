# CQRS Handler Testing Guide

This guide explains how to write unit tests for CQRS handlers in the ToxicToastGo microservices.

## Overview

All services using CQRS pattern should have comprehensive unit tests for:
- **Command Validation**: Ensure commands validate input correctly
- **Command Handlers**: Test business logic and error scenarios
- **Query Validation**: Ensure queries validate parameters
- **Query Handlers**: Test data retrieval and transformation

## Test Structure

### Directory Layout

```
services/my-service/
├── internal/
│   ├── command/
│   │   ├── my_commands.go
│   │   └── my_commands_test.go      # ← Command tests
│   └── query/
│       ├── my_queries.go
│       └── my_queries_test.go        # ← Query tests
```

### Test File Naming

- Tests must be in the same package as the code they test
- Test file name: `<original_file>_test.go`
- Example: `post_commands.go` → `post_commands_test.go`

## Command Handler Tests

### Example: blog-service

See `services/blog-service/internal/command/post_commands_test.go` for a complete example.

### Test Template

```go
package command

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/MY-SERVICE/internal/domain"
	"toxictoast/services/MY-SERVICE/internal/repository"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type MockMyRepository struct {
	mock.Mock
}

func (m *MockMyRepository) Create(ctx context.Context, entity *domain.MyEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockMyRepository) GetByID(ctx context.Context, id string) (*domain.MyEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MyEntity), args.Error(1)
}

// ... implement all repository methods

// ============================================================================
// Command Validation Tests
// ============================================================================

func TestCreateMyEntityCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *CreateMyEntityCommand
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: &CreateMyEntityCommand{
				Name: "Test Entity",
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			cmd: &CreateMyEntityCommand{
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Command Handler Tests
// ============================================================================

func TestCreateMyEntityHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockMyRepository)

		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.MyEntity")).Return(nil)

		handler := NewCreateMyEntityHandler(mockRepo)

		cmd := &CreateMyEntityCommand{
			Name: "Test Entity",
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		assert.NotEmpty(t, cmd.AggregateID) // Ensure UUID was generated
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockMyRepository)

		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.MyEntity")).
			Return(errors.New("database error"))

		handler := NewCreateMyEntityHandler(mockRepo)

		cmd := &CreateMyEntityCommand{
			Name: "Test Entity",
		}

		err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create")
	})
}
```

### Key Patterns

1. **Mock All Repository Dependencies**
   ```go
   type MockMyRepository struct {
       mock.Mock
   }
   ```

2. **Test Both Success and Error Scenarios**
   - Happy path: Command succeeds
   - Repository errors
   - Validation errors
   - Business rule violations

3. **Verify Side Effects**
   ```go
   assert.NotEmpty(t, cmd.AggregateID) // UUID generation
   mockRepo.AssertExpectations(t)      // Correct calls made
   ```

4. **Use Table-Driven Tests for Validation**
   ```go
   tests := []struct {
       name    string
       cmd     *Command
       wantErr bool
   }{...}
   ```

## Query Handler Tests

### Example: blog-service

See `services/blog-service/internal/query/post_queries_test.go` for a complete example.

### Test Template

```go
package query

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"toxictoast/services/MY-SERVICE/internal/domain"
	"toxictoast/services/MY-SERVICE/internal/repository"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockMyRepository struct {
	mock.Mock
}

func (m *MockMyRepository) GetByID(ctx context.Context, id string) (*domain.MyEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MyEntity), args.Error(1)
}

func (m *MockMyRepository) List(ctx context.Context, filters repository.MyFilters) ([]domain.MyEntity, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.MyEntity), args.Get(1).(int64), args.Error(2)
}

// ============================================================================
// Query Validation Tests
// ============================================================================

func TestGetMyEntityByIDQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *GetMyEntityByIDQuery
		wantErr bool
	}{
		{
			name: "valid query",
			query: &GetMyEntityByIDQuery{
				ID: "entity-123",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			query: &GetMyEntityByIDQuery{
				ID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Query Handler Tests
// ============================================================================

func TestGetMyEntityByIDHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockMyRepository)

		expectedEntity := &domain.MyEntity{
			ID:   "entity-123",
			Name: "Test Entity",
		}

		mockRepo.On("GetByID", ctx, "entity-123").Return(expectedEntity, nil)

		handler := NewGetMyEntityByIDHandler(mockRepo)

		query := &GetMyEntityByIDQuery{
			ID: "entity-123",
		}

		result, err := handler.Handle(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		entity := result.(*domain.MyEntity)
		assert.Equal(t, "entity-123", entity.ID)
		assert.Equal(t, "Test Entity", entity.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("entity not found", func(t *testing.T) {
		mockRepo := new(MockMyRepository)

		mockRepo.On("GetByID", ctx, "non-existent").Return(nil, errors.New("not found"))

		handler := NewGetMyEntityByIDHandler(mockRepo)

		query := &GetMyEntityByIDQuery{
			ID: "non-existent",
		}

		result, err := handler.Handle(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get")
	})
}

func TestListMyEntitiesHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful listing", func(t *testing.T) {
		mockRepo := new(MockMyRepository)

		expectedEntities := []domain.MyEntity{
			{ID: "entity-1", Name: "Entity 1"},
			{ID: "entity-2", Name: "Entity 2"},
		}

		filters := repository.MyFilters{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("List", ctx, filters).Return(expectedEntities, int64(2), nil)

		handler := NewListMyEntitiesHandler(mockRepo)

		query := &ListMyEntitiesQuery{
			Filters: filters,
		}

		result, err := handler.Handle(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		listResult := result.(*ListMyEntitiesResult)
		assert.Len(t, listResult.Entities, 2)
		assert.Equal(t, int64(2), listResult.Total)

		mockRepo.AssertExpectations(t)
	})
}
```

## Special Cases

### Query-Only Services (e.g., weather-service)

For services that only fetch data from external APIs:

```go
// Focus on validation tests only
func TestGetWeatherQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *GetWeatherQuery
		wantErr bool
	}{
		{
			name: "valid coordinates",
			query: &GetWeatherQuery{
				Latitude:  52.52,
				Longitude: 13.405,
			},
			wantErr: false,
		},
		{
			name: "invalid latitude",
			query: &GetWeatherQuery{
				Latitude:  91, // > 90
				Longitude: 13.405,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

**Note**: Handler tests for external APIs require HTTP mocking (e.g., `httptest`) which is more complex and often better suited for integration tests.

### Commands with Kafka Events

When commands publish Kafka events:

```go
func TestCreateEntityHandler_Handle(t *testing.T) {
	t.Run("successful creation with kafka event", func(t *testing.T) {
		mockRepo := new(MockEntityRepository)
		// Note: kafkaProducer passed as nil in tests
		// Handler should check if kafkaProducer != nil before publishing

		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Entity")).Return(nil)

		handler := NewCreateEntityHandler(mockRepo, nil) // nil kafkaProducer

		cmd := &CreateEntityCommand{
			Name: "Test",
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		// Event publishing is optional (logged warning if fails)
		mockRepo.AssertExpectations(t)
	})
}
```

## Running Tests

### Run Tests for Single Service

```bash
cd services/my-service
go test ./internal/command -v
go test ./internal/query -v
```

### Run All Tests with Coverage

```bash
cd services/my-service
go test ./internal/... -v -cover
```

### Expected Coverage

- **Validation Logic**: Aim for 100% coverage
- **Handler Logic**: Aim for 70-90% coverage
- **Overall CQRS Layer**: 15-25% is typical (due to infrastructure code)

## Test Checklist

When writing tests for a new service:

- [ ] Mock all repository interfaces
- [ ] Test command validation (all required fields, boundaries)
- [ ] Test command handler success scenarios
- [ ] Test command handler error scenarios
- [ ] Test query validation
- [ ] Test query handler success scenarios
- [ ] Test query handler not found scenarios
- [ ] Test query handler listing with pagination
- [ ] Verify all mocks with `AssertExpectations(t)`
- [ ] Run tests with `-v` flag to see all test names
- [ ] Run tests with `-cover` to check coverage

## Dependencies

All test files require:

```go
import (
	"github.com/stretchr/testify/assert"  // Assertions
	"github.com/stretchr/testify/mock"    // Mocking
)
```

Install with:

```bash
go get github.com/stretchr/testify
```

## Examples by Service

### Completed Examples

✅ **blog-service**: Full command + query tests (991 lines)
- Location: `services/blog-service/internal/{command,query}/post_*_test.go`
- Coverage: 16.3% (commands), 20.0% (queries)
- Tests: 24 tests, 57 subtests

✅ **weather-service**: Query validation tests (202 lines)
- Location: `services/weather-service/internal/query/weather_queries_test.go`
- Tests: 4 tests, 16 subtests
- Focus: Input validation (external API integration tested separately)

### Services Needing Tests

The following services have CQRS handlers but no tests yet:

1. **auth-service** - 3 commands, 2 queries (JWT authentication)
2. **foodfolio-service** - 14 commands, 14 queries (inventory management)
3. **link-service** - 3 commands, 4 queries (URL shortener)
4. **notification-service** - 4 commands, 5 queries (Discord webhooks)
5. **twitchbot-service** - 19 commands, 21 queries (Twitch bot management)
6. **webhook-service** - 4 commands, 4 queries (webhook delivery)
7. **warcraft-service** - 0 commands, 10 queries (Battle.net API)
8. **user-service** - Needs CQRS migration first
9. **sse-service** - Needs CQRS migration first
10. **gateway-service** - Gateway, not domain logic

## Best Practices

### 1. One Mock Per Interface

```go
// Good: Separate mocks
type MockPostRepository struct { mock.Mock }
type MockCategoryRepository struct { mock.Mock }

// Bad: One mock for everything (hard to maintain)
type MockAllRepositories struct { mock.Mock }
```

### 2. Use Table-Driven Tests

```go
tests := []struct {
    name    string
    input   Input
    wantErr bool
}{
    {"valid input", validInput, false},
    {"invalid input", invalidInput, true},
}
```

### 3. Test Error Messages

```go
assert.Error(t, err)
assert.Contains(t, err.Error(), "expected error substring")
```

### 4. Verify Mock Calls

```go
mockRepo.AssertExpectations(t) // All expected calls made
mockRepo.AssertNotCalled(t, "MethodName") // Method not called
```

### 5. Use Descriptive Test Names

```go
// Good
t.Run("successful creation with categories", func(t *testing.T) { ... })

// Bad
t.Run("test1", func(t *testing.T) { ... })
```

## Troubleshooting

### "cannot use mockRepo as repository.XRepository"

Ensure your mock implements ALL interface methods. Check the repository interface definition and add missing methods to your mock.

### "args.Get(0).(type) panics"

Always check for nil before type assertion:

```go
if args.Get(0) == nil {
    return nil, args.Error(1)
}
return args.Get(0).(*domain.Entity), args.Error(1)
```

### "undefined: cqrs.BaseCommand"

Import the shared CQRS package:

```go
import "github.com/toxictoast/toxictoastgo/shared/cqrs"
```

### Filter Struct Field Mismatch

Ensure filter structs use correct field names from repository interface:

```go
// Repository uses:
type PostFilters struct {
    Page     int
    PageSize int
}

// Test should use:
filters := repository.PostFilters{
    Page:     1,
    PageSize: 10,
}
```

## Contributing

When adding tests for a new service:

1. Follow the templates above
2. Run tests locally: `go test ./internal/... -v -cover`
3. Ensure all tests pass
4. Commit with message: `test(service-name): add CQRS handler tests`
5. Include coverage stats in commit message

---

**Last Updated**: 2025-11-28
**Test Coverage Status**: 2/12 services have CQRS tests
