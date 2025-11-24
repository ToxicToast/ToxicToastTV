# Event Sourcing & CQRS

Comprehensive Event Sourcing and CQRS implementation for ToxicToastGo microservices.

## Overview

This package provides the infrastructure for:
- **Event Sourcing**: Storing all changes as immutable events
- **CQRS**: Separating reads (queries) from writes (commands)
- **Event Store**: PostgreSQL-based persistent event storage
- **Aggregates**: Domain objects reconstructed from events
- **Projectors**: Building read models from events

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Command Side (Write)                     │
├─────────────────────────────────────────────────────────────┤
│  Command → CommandHandler → Aggregate → Events → EventStore │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Events
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Query Side (Read)                        │
├─────────────────────────────────────────────────────────────┤
│    Events → Projector → ReadModel → QueryHandler → Query    │
└─────────────────────────────────────────────────────────────┘
```

## Core Concepts

### Event Sourcing

Instead of storing current state, we store **all changes as events**:

```go
// Traditional: Store current state
User{ ID: "123", Name: "John", Email: "john@example.com" }

// Event Sourcing: Store all changes
UserCreatedEvent{ ID: "123", Name: "John", Email: "john@example.com" }
UserEmailChangedEvent{ ID: "123", OldEmail: "john@example.com", NewEmail: "john@new.com" }
UserNameChangedEvent{ ID: "123", OldName: "John", NewName: "John Doe" }
```

**Benefits:**
- Complete audit trail
- Time travel (reconstruct state at any point)
- Easy debugging (replay events)
- Event replay for new features
- No data loss

### CQRS (Command Query Responsibility Segregation)

Separate models for **writes** (commands) and **reads** (queries):

```go
// Write Side: Commands modify aggregates
CreateUserCommand → UserAggregate → UserCreatedEvent

// Read Side: Queries read from optimized models
GetUserQuery → UserReadModel → User DTO
```

**Benefits:**
- Optimized read models
- Scale reads and writes independently
- Different consistency models (eventual consistency)
- Multiple read models from same events

## Event Store

### Event Envelope

All events are wrapped in an envelope with metadata:

```go
type EventEnvelope struct {
    EventID       string          // Unique event ID
    EventType     string          // e.g., "user.created"
    AggregateID   string          // ID of the aggregate
    AggregateType string          // e.g., "user"
    Version       int64           // Aggregate version (for optimistic locking)
    Timestamp     time.Time       // When the event occurred
    Data          json.RawMessage // Event payload
    Metadata      map[string]interface{} // Additional metadata
}
```

### PostgreSQL Event Store

```go
// Create event store
db, _ := sql.Open("postgres", connectionString)
eventStore, err := eventstore.NewPostgresEventStore(db)

// Save events with optimistic locking
events := []*EventEnvelope{userCreatedEvent}
err = eventStore.SaveEvents(ctx, aggregateID, expectedVersion, events)

// Load events for an aggregate
events, err := eventStore.GetEvents(ctx, "user", userID)

// Get event stream for projections
events, err := eventStore.GetEventStream(ctx, sinceTimestamp, limit)
```

### Database Schema

```sql
CREATE TABLE event_store (
    event_id VARCHAR(36) PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(36) NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    version BIGINT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    data JSONB NOT NULL,
    metadata JSONB,
    UNIQUE(aggregate_id, version)
);
```

## Aggregates

### Base Aggregate

```go
type UserAggregate struct {
    *eventstore.BaseAggregate
    Email    string
    Username string
    IsActive bool
}

func NewUserAggregate(id string) *UserAggregate {
    return &UserAggregate{
        BaseAggregate: eventstore.NewBaseAggregate(id, eventstore.AggregateTypeUser),
    }
}
```

### Raising Events

```go
func (a *UserAggregate) CreateUser(email, username string) error {
    event := UserCreatedEvent{
        Email:    email,
        Username: username,
    }

    return a.RaiseEvent(eventstore.EventTypeUserCreated, event, func(e *eventstore.EventEnvelope) error {
        return a.applyUserCreated(e)
    })
}

func (a *UserAggregate) applyUserCreated(e *eventstore.EventEnvelope) error {
    var event UserCreatedEvent
    if err := e.UnmarshalData(&event); err != nil {
        return err
    }

    a.Email = event.Email
    a.Username = event.Username
    a.IsActive = true
    return nil
}
```

### Loading from History

```go
func (a *UserAggregate) LoadFromHistory(events []*eventstore.EventEnvelope) error {
    return a.BaseAggregate.LoadFromHistory(events, func(e *eventstore.EventEnvelope) error {
        switch e.EventType {
        case eventstore.EventTypeUserCreated:
            return a.applyUserCreated(e)
        case eventstore.EventTypeUserUpdated:
            return a.applyUserUpdated(e)
        default:
            return fmt.Errorf("unknown event type: %s", e.EventType)
        }
    })
}
```

### Aggregate Repository

```go
repo := eventstore.NewAggregateRepository(eventStore)

// Load aggregate
user := NewUserAggregate(userID)
err := repo.Load(ctx, user)

// Modify aggregate
user.ChangeEmail("new@example.com")

// Save changes
err = repo.Save(ctx, user)
```

## CQRS

### Commands

```go
// Define command
type CreateUserCommand struct {
    cqrs.BaseCommand
    Email    string
    Username string
}

func (c *CreateUserCommand) CommandName() string {
    return "create_user"
}

func (c *CreateUserCommand) Validate() error {
    if c.Email == "" {
        return errors.New("email is required")
    }
    return nil
}

// Command handler
type CreateUserHandler struct {
    repo *eventstore.AggregateRepository
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
    createCmd := cmd.(*CreateUserCommand)

    user := NewUserAggregate(createCmd.AggregateID)
    if err := user.CreateUser(createCmd.Email, createCmd.Username); err != nil {
        return err
    }

    return h.repo.Save(ctx, user)
}

// Register and dispatch
commandBus := cqrs.NewCommandBus()
commandBus.RegisterHandler("create_user", &CreateUserHandler{repo: repo})

cmd := &CreateUserCommand{
    BaseCommand: cqrs.BaseCommand{AggregateID: userID},
    Email:       "user@example.com",
    Username:    "john_doe",
}

err := commandBus.Dispatch(ctx, cmd)
```

### Queries

```go
// Define query
type GetUserQuery struct {
    cqrs.BaseQuery
    UserID string
}

func (q *GetUserQuery) QueryName() string {
    return "get_user"
}

// Query handler
type GetUserHandler struct {
    readModelRepo ReadModelRepository
}

func (h *GetUserHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
    q := query.(*GetUserQuery)
    return h.readModelRepo.FindByID(ctx, q.UserID)
}

// Register and dispatch
queryBus := cqrs.NewQueryBus()
queryBus.RegisterHandler("get_user", &GetUserHandler{readModelRepo: repo})

query := &GetUserQuery{UserID: userID}
result, err := queryBus.Dispatch(ctx, query)
user := result.(*UserReadModel)
```

### Read Models

```go
// Define read model
type UserReadModel struct {
    *cqrs.BaseReadModel
    Email     string    `json:"email" db:"email"`
    Username  string    `json:"username" db:"username"`
    IsActive  bool      `json:"is_active" db:"is_active"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Read model repository
type UserReadModelRepository struct {
    db *sql.DB
}

func (r *UserReadModelRepository) FindByID(ctx context.Context, id string) (cqrs.ReadModel, error) {
    var user UserReadModel
    err := r.db.QueryRowContext(ctx,
        "SELECT id, email, username, is_active, created_at, last_updated FROM user_read_model WHERE id = $1",
        id,
    ).Scan(&user.ID, &user.Email, &user.Username, &user.IsActive, &user.CreatedAt, &user.LastUpdated)

    return &user, err
}

func (r *UserReadModelRepository) Save(ctx context.Context, model cqrs.ReadModel) error {
    user := model.(*UserReadModel)
    _, err := r.db.ExecContext(ctx, `
        INSERT INTO user_read_model (id, email, username, is_active, created_at, last_updated)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE SET
            email = EXCLUDED.email,
            username = EXCLUDED.username,
            is_active = EXCLUDED.is_active,
            last_updated = EXCLUDED.last_updated
    `, user.ID, user.Email, user.Username, user.IsActive, user.CreatedAt, user.LastUpdated)

    return err
}
```

### Projectors

```go
// Define projector
type UserProjector struct {
    readModelRepo *UserReadModelRepository
}

func (p *UserProjector) GetProjectorName() string {
    return "user_projector"
}

func (p *UserProjector) GetEventTypes() []string {
    return []string{
        eventstore.EventTypeUserCreated,
        eventstore.EventTypeUserUpdated,
    }
}

func (p *UserProjector) ProjectEvent(ctx context.Context, event *eventstore.EventEnvelope) error {
    switch event.EventType {
    case eventstore.EventTypeUserCreated:
        return p.projectUserCreated(ctx, event)
    case eventstore.EventTypeUserUpdated:
        return p.projectUserUpdated(ctx, event)
    }
    return nil
}

func (p *UserProjector) projectUserCreated(ctx context.Context, event *eventstore.EventEnvelope) error {
    var e UserCreatedEvent
    if err := event.UnmarshalData(&e); err != nil {
        return err
    }

    user := &UserReadModel{
        BaseReadModel: cqrs.NewBaseReadModel(event.AggregateID),
        Email:        e.Email,
        Username:     e.Username,
        IsActive:     true,
        CreatedAt:    event.Timestamp,
    }

    return p.readModelRepo.Save(ctx, user)
}

// Register projector
projectorManager := cqrs.NewProjectorManager(eventStore)
projectorManager.RegisterProjector(&UserProjector{readModelRepo: repo})

// Project event
projectorManager.ProjectEvent(ctx, event)

// Rebuild all projections
projectorManager.RebuildProjections(ctx, eventstore.AggregateTypeUser)
```

## Integration Example

```go
// 1. Setup
db, _ := sql.Open("postgres", connStr)
eventStore, _ := eventstore.NewPostgresEventStore(db)
repo := eventstore.NewAggregateRepository(eventStore)

// 2. Command Side
commandBus := cqrs.NewCommandBus()
commandBus.RegisterHandler("create_user", &CreateUserHandler{repo: repo})

cmd := &CreateUserCommand{
    BaseCommand: cqrs.BaseCommand{AggregateID: uuid.New().String()},
    Email:       "user@example.com",
    Username:    "johndoe",
}

err := commandBus.Dispatch(ctx, cmd)

// 3. Event is persisted
// UserCreatedEvent → EventStore

// 4. Projector builds read model
projectorManager := cqrs.NewProjectorManager(eventStore)
projectorManager.RegisterProjector(&UserProjector{readModelRepo: userReadRepo})

// 5. Query Side
queryBus := cqrs.NewQueryBus()
queryBus.RegisterHandler("get_user", &GetUserHandler{readModelRepo: userReadRepo})

query := &GetUserQuery{UserID: cmd.AggregateID}
result, _ := queryBus.Dispatch(ctx, query)
user := result.(*UserReadModel)
```

## Best Practices

### 1. Event Design
- Events are immutable
- Events represent past facts (UserCreated, not CreateUser)
- Include all data needed to rebuild state
- Version events for schema evolution

### 2. Aggregate Design
- Keep aggregates small and focused
- One aggregate per transaction
- Validate invariants before raising events
- Use eventual consistency between aggregates

### 3. Read Model Design
- Optimize for query patterns
- Denormalize as needed
- Multiple read models from same events
- Eventually consistent with write side

### 4. Error Handling
- Handle concurrency conflicts (optimistic locking)
- Implement retry logic for projections
- Monitor projection lag
- Handle poison messages

### 5. Testing
- Test aggregates with given/when/then
- Test projectors with event sequences
- Test concurrency scenarios
- Integration tests with real event store

## Migration Strategy

1. **Add Event Store alongside existing repository**
2. **Dual-write to both stores initially**
3. **Build projectors for read models**
4. **Switch reads to query new read models**
5. **Remove old repository once validated**

## Performance Considerations

- Use snapshots for large aggregates (100+ events)
- Batch event projections
- Cache read models
- Partition event store by aggregate type
- Use JSONB indexes for event data queries

## Monitoring

- Track event store size
- Monitor projection lag
- Alert on concurrency conflicts
- Track command/query latency
- Monitor read model freshness

## License

Proprietary - ToxicToast
