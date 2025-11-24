# User Service - CQRS Migration Guide

Complete CQRS and Event Sourcing implementation for the User Service.

## Overview

The User Service now implements **Event Sourcing** and **CQRS** patterns:

- **Event Sourcing**: All user changes are stored as immutable events
- **CQRS**: Separate models for writes (commands) and reads (queries)
- **Event Store**: PostgreSQL-based persistent event storage
- **Read Models**: Optimized projections for fast queries
- **Audit Trail**: Complete history of all user changes

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  COMMAND SIDE (Write)               │
├─────────────────────────────────────────────────────┤
│                                                     │
│  gRPC Request                                       │
│      │                                              │
│      ▼                                              │
│  Command (CreateUser, ChangeEmail, etc.)            │
│      │                                              │
│      ▼                                              │
│  CommandHandler                                     │
│      │                                              │
│      ▼                                              │
│  UserAggregate (Business Logic)                     │
│      │                                              │
│      ▼                                              │
│  Events (UserCreated, EmailChanged, etc.)           │
│      │                                              │
│      ▼                                              │
│  EventStore (PostgreSQL)                            │
│                                                     │
└─────────────────────────────────────────────────────┘
                      │
                      │ Events
                      ▼
┌─────────────────────────────────────────────────────┐
│                  QUERY SIDE (Read)                  │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Events                                             │
│      │                                              │
│      ▼                                              │
│  UserProjector                                      │
│      │                                              │
│      ▼                                              │
│  UserReadModel (PostgreSQL)                         │
│      │                                              │
│      ▼                                              │
│  QueryHandler                                       │
│      │                                              │
│      ▼                                              │
│  gRPC Response                                      │
│                                                     │
└─────────────────────────────────────────────────────┘
```

## New Components

### 1. User Aggregate (`internal/aggregate/user_aggregate.go`)

Event-sourced aggregate that encapsulates user business logic:

```go
type UserAggregate struct {
    *eventstore.BaseAggregate
    Email        string
    Username     string
    PasswordHash string
    Status       domain.UserStatus
    // ... other fields
}
```

**Methods:**
- `CreateUser()` - Creates new user and raises UserCreatedEvent
- `ChangeEmail()` - Changes email and raises EmailChangedEvent
- `ChangePassword()` - Changes password and raises PasswordChangedEvent
- `UpdateProfile()` - Updates profile and raises ProfileUpdatedEvent
- `Activate()` / `Deactivate()` - Changes user status
- `Delete()` - Soft deletes user
- `LoadFromHistory()` - Reconstructs state from events

### 2. Commands (`internal/command/user_commands.go`)

Commands represent intentions to change user state:

**Available Commands:**
- `CreateUserCommand` - Create new user
- `ChangeEmailCommand` - Change user email
- `ChangePasswordCommand` - Change user password
- `UpdateProfileCommand` - Update profile (firstName, lastName, avatarURL)
- `ActivateUserCommand` - Activate user
- `DeactivateUserCommand` - Deactivate user
- `DeleteUserCommand` - Soft delete user

**Command Handlers:**
- Handle business logic
- Validate commands
- Load/Save aggregates
- Check business rules (uniqueness, etc.)

### 3. Events

All user changes are captured as events:

```go
UserCreatedEvent {
    UserID, Email, Username, PasswordHash,
    FirstName, LastName, AvatarURL, Status, CreatedAt
}

UserEmailChangedEvent {
    UserID, OldEmail, NewEmail, UpdatedAt
}

UserPasswordChangedEvent {
    UserID, NewPasswordHash, UpdatedAt
}

UserProfileUpdatedEvent {
    UserID, FirstName, LastName, AvatarURL, UpdatedAt
}

UserActivatedEvent / UserDeactivatedEvent / UserDeletedEvent
```

### 4. Read Model (`internal/projection/user_read_model.go`)

Optimized projection for fast queries:

```sql
CREATE TABLE user_read_model (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    avatar_url TEXT,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);
```

**Query Methods:**
- `FindByID(id)` - Find by user ID
- `FindByEmail(email)` - Find by email
- `FindByUsername(username)` - Find by username
- `FindAll(limit, offset)` - List users with pagination

### 5. Projector

Projects events into read models:

```go
type UserProjector struct {
    repo *UserReadModelRepository
}

// Handles these events:
- UserCreatedEvent → Create read model
- UserEmailChangedEvent → Update email
- UserProfileUpdatedEvent → Update profile
- UserActivatedEvent → Update status
- UserDeactivatedEvent → Update status
- UserDeletedEvent → Soft delete
```

## Integration Steps

### Step 1: Setup Event Store

```go
// In main.go or service setup
db, err := sql.Open("postgres", connectionString)
eventStore, err := eventstore.NewPostgresEventStore(db)
aggRepo := eventstore.NewAggregateRepository(eventStore)
```

### Step 2: Setup Command Bus

```go
commandBus := cqrs.NewCommandBus()

// Register all command handlers
commandBus.RegisterHandler("create_user",
    command.NewCreateUserHandler(aggRepo, userRepo))
commandBus.RegisterHandler("change_email",
    command.NewChangeEmailHandler(aggRepo, userRepo))
commandBus.RegisterHandler("change_password",
    command.NewChangePasswordHandler(aggRepo))
// ... register all handlers
```

### Step 3: Setup Projector

```go
// Create read model repository
readModelRepo := projection.NewUserReadModelRepository(db)
readModelRepo.CreateTable() // Create read model table

// Create projector
userProjector := projection.NewUserProjector(readModelRepo)

// Register with projector manager
projectorManager := cqrs.NewProjectorManager(eventStore)
projectorManager.RegisterProjector(userProjector)

// Optional: Rebuild projections from history
projectorManager.RebuildProjections(ctx, eventstore.AggregateTypeUser)
```

### Step 4: Use in gRPC Handler

**Write Operation (Create User):**

```go
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    cmd := &command.CreateUserCommand{
        BaseCommand: cqrs.BaseCommand{},
        Email:       req.Email,
        Username:    req.Username,
        Password:    req.Password,
        FirstName:   &req.FirstName,
        LastName:    &req.LastName,
        AvatarURL:   &req.AvatarUrl,
    }

    // Dispatch command
    if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
        return nil, err
    }

    // Read from read model
    user, err := h.readModelRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }

    return toProtoUser(user), nil
}
```

**Read Operation (Get User):**

```go
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Query read model directly (fast!)
    user, err := h.readModelRepo.FindByID(ctx, req.Id)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, status.Error(codes.NotFound, "user not found")
    }

    return toProtoUser(user), nil
}
```

## Migration Strategy

### Phase 1: Dual-Write (Current + Event Store)

```go
// Keep existing implementation
user, err := uc.CreateUser(ctx, email, username, password, ...)

// ALSO write to event store
cmd := &command.CreateUserCommand{...}
commandBus.Dispatch(ctx, cmd)
```

### Phase 2: Validate Read Models

```go
// Compare old vs new
oldUser, _ := oldRepo.GetByID(ctx, id)
newUser, _ := readModelRepo.FindByID(ctx, id)

// Validate they match
if !compareUsers(oldUser, newUser) {
    log.Error("Read model mismatch!")
}
```

### Phase 3: Switch Reads to Read Models

```go
// Change gRPC handlers to use read models
user, err := h.readModelRepo.FindByEmail(ctx, email)
```

### Phase 4: Switch Writes to Commands

```go
// Use command bus for all writes
err := h.commandBus.Dispatch(ctx, cmd)
```

### Phase 5: Remove Old Repository

Once validated, remove the old CRUD repository.

## Benefits

### 1. Complete Audit Trail
```go
// Get all events for a user
events, _ := eventStore.GetEvents(ctx, "user", userID)

// See full history:
// - UserCreated at 2025-01-01 10:00
// - EmailChanged from old@test.com to new@test.com at 2025-01-02 15:30
// - PasswordChanged at 2025-01-03 09:15
// - UserDeactivated at 2025-01-05 14:00
```

### 2. Time Travel
```go
// Reconstruct user state at any point in time
user := aggregate.NewUserAggregate(userID)
eventsUntil := getEventsUntil(ctx, userID, "2025-01-03")
user.LoadFromHistory(eventsUntil)
// User state as it was on 2025-01-03
```

### 3. Event Replay
```go
// Add new projections from existing events
projectorManager.RegisterProjector(newProjector)
projectorManager.RebuildProjections(ctx, "user")
// New read model built from all historical events!
```

### 4. Better Performance
- Writes: Append-only (fast)
- Reads: Optimized read models (no joins)
- Scale reads and writes independently

### 5. Debugging
```go
// Replay events to debug issues
for _, event := range events {
    log.Printf("Event: %s at %s", event.EventType, event.Timestamp)
    log.Printf("Data: %s", event.Data)
}
```

## Testing

### Test Aggregate

```go
func TestUserAggregate_CreateUser(t *testing.T) {
    user := aggregate.NewUserAggregate("test-id")

    err := user.CreateUser("test@example.com", "testuser", "hash", nil, nil, nil)
    assert.NoError(t, err)

    events := user.GetUncommittedEvents()
    assert.Len(t, events, 1)
    assert.Equal(t, eventstore.EventTypeUserCreated, events[0].EventType)
}
```

### Test Command Handler

```go
func TestCreateUserHandler(t *testing.T) {
    handler := command.NewCreateUserHandler(aggRepo, userRepo)

    cmd := &command.CreateUserCommand{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password",
    }

    err := handler.Handle(ctx, cmd)
    assert.NoError(t, err)
}
```

### Test Projector

```go
func TestUserProjector(t *testing.T) {
    projector := projection.NewUserProjector(readModelRepo)

    event := createUserCreatedEvent()
    err := projector.ProjectEvent(ctx, event)
    assert.NoError(t, err)

    user, _ := readModelRepo.FindByID(ctx, event.AggregateID)
    assert.NotNil(t, user)
}
```

## Monitoring

### Event Store Metrics
- Total events count
- Events per aggregate type
- Event store size
- Write throughput

### Projection Lag
- Time between event creation and projection
- Failed projections count
- Retry attempts

### Query Performance
- Read model query times
- Cache hit rates
- Most queried users

## Troubleshooting

### Projection Lag
```go
// Check last processed event
lastEvent := getLastProjectedEvent()
currentEvent := getLatestEvent()

lag := currentEvent.Timestamp - lastEvent.Timestamp
if lag > threshold {
    log.Warning("Projection lag detected: %v", lag)
}
```

### Concurrency Conflicts
```go
// Retry on concurrency conflicts
for retries := 0; retries < maxRetries; retries++ {
    if err := aggRepo.Save(ctx, user); err == eventstore.ErrConcurrencyConflict {
        user = aggregate.NewUserAggregate(userID)
        aggRepo.Load(ctx, user)
        continue
    }
    break
}
```

### Read Model Inconsistency
```go
// Rebuild single user projection
events, _ := eventStore.GetEvents(ctx, "user", userID)
for _, event := range events {
    projector.ProjectEvent(ctx, event)
}
```

## Next Steps

1. ✅ User Service CQRS implemented
2. ⏳ Add integration tests
3. ⏳ Implement Auth Service with CQRS
4. ⏳ Add monitoring dashboards
5. ⏳ Performance benchmarking

## License

Proprietary - ToxicToast
