package aggregate

import (
	"errors"
	"fmt"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/eventstore"
	"toxictoast/services/user-service/internal/domain"
)

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidUsername       = errors.New("invalid username")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrUserNotActive         = errors.New("user is not active")
	ErrUserAlreadyActive     = errors.New("user is already active")
	ErrUserAlreadyDeleted    = errors.New("user is already deleted")
)

// UserAggregate represents the user aggregate root
type UserAggregate struct {
	*eventstore.BaseAggregate

	// Current state
	Email        string
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarURL    string
	Status       domain.UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// NewUserAggregate creates a new user aggregate
func NewUserAggregate(id string) *UserAggregate {
	return &UserAggregate{
		BaseAggregate: eventstore.NewBaseAggregate(id, eventstore.AggregateTypeUser),
	}
}

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	UserID       string             `json:"user_id"`
	Email        string             `json:"email"`
	Username     string             `json:"username"`
	PasswordHash string             `json:"password_hash"`
	FirstName    string             `json:"first_name,omitempty"`
	LastName     string             `json:"last_name,omitempty"`
	AvatarURL    string             `json:"avatar_url,omitempty"`
	Status       domain.UserStatus  `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
}

// UserEmailChangedEvent represents an email change event
type UserEmailChangedEvent struct {
	UserID   string    `json:"user_id"`
	OldEmail string    `json:"old_email"`
	NewEmail string    `json:"new_email"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPasswordChangedEvent represents a password change event
type UserPasswordChangedEvent struct {
	UserID          string    `json:"user_id"`
	NewPasswordHash string    `json:"new_password_hash"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserProfileUpdatedEvent represents a profile update event
type UserProfileUpdatedEvent struct {
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserActivatedEvent represents user activation
type UserActivatedEvent struct {
	UserID      string    `json:"user_id"`
	ActivatedAt time.Time `json:"activated_at"`
}

// UserDeactivatedEvent represents user deactivation
type UserDeactivatedEvent struct {
	UserID        string    `json:"user_id"`
	DeactivatedAt time.Time `json:"deactivated_at"`
}

// UserDeletedEvent represents user deletion (soft delete)
type UserDeletedEvent struct {
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// CreateUser creates a new user
func (a *UserAggregate) CreateUser(email, username, passwordHash string, firstName, lastName, avatarURL *string) error {
	// Validation
	if email == "" {
		return ErrInvalidEmail
	}
	if username == "" {
		return ErrInvalidUsername
	}
	if passwordHash == "" {
		return ErrInvalidPassword
	}

	// Create event
	event := UserCreatedEvent{
		UserID:       a.ID,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Status:       domain.UserStatusActive,
		CreatedAt:    time.Now().UTC(),
	}

	if firstName != nil {
		event.FirstName = *firstName
	}
	if lastName != nil {
		event.LastName = *lastName
	}
	if avatarURL != nil {
		event.AvatarURL = *avatarURL
	}

	return a.RaiseEvent(eventstore.EventTypeUserCreated, event, func(e *eventstore.EventEnvelope) error {
		return a.applyUserCreated(e)
	})
}

// ChangeEmail changes the user's email
func (a *UserAggregate) ChangeEmail(newEmail string) error {
	if newEmail == "" {
		return ErrInvalidEmail
	}

	if a.Email == newEmail {
		return nil // No change
	}

	event := UserEmailChangedEvent{
		UserID:    a.ID,
		OldEmail:  a.Email,
		NewEmail:  newEmail,
		UpdatedAt: time.Now().UTC(),
	}

	return a.RaiseEvent(eventstore.EventTypeUserUpdated, event, func(e *eventstore.EventEnvelope) error {
		return a.applyEmailChanged(e)
	})
}

// ChangePassword changes the user's password
func (a *UserAggregate) ChangePassword(newPasswordHash string) error {
	if newPasswordHash == "" {
		return ErrInvalidPassword
	}

	event := UserPasswordChangedEvent{
		UserID:          a.ID,
		NewPasswordHash: newPasswordHash,
		UpdatedAt:       time.Now().UTC(),
	}

	return a.RaiseEvent(eventstore.EventTypeUserUpdated, event, func(e *eventstore.EventEnvelope) error {
		return a.applyPasswordChanged(e)
	})
}

// UpdateProfile updates the user's profile information
func (a *UserAggregate) UpdateProfile(firstName, lastName, avatarURL *string) error {
	event := UserProfileUpdatedEvent{
		UserID:    a.ID,
		UpdatedAt: time.Now().UTC(),
	}

	if firstName != nil {
		event.FirstName = *firstName
	}
	if lastName != nil {
		event.LastName = *lastName
	}
	if avatarURL != nil {
		event.AvatarURL = *avatarURL
	}

	return a.RaiseEvent(eventstore.EventTypeUserUpdated, event, func(e *eventstore.EventEnvelope) error {
		return a.applyProfileUpdated(e)
	})
}

// Activate activates the user
func (a *UserAggregate) Activate() error {
	if a.Status == domain.UserStatusActive {
		return ErrUserAlreadyActive
	}

	event := UserActivatedEvent{
		UserID:      a.ID,
		ActivatedAt: time.Now().UTC(),
	}

	return a.RaiseEvent(eventstore.EventTypeUserActivated, event, func(e *eventstore.EventEnvelope) error {
		return a.applyUserActivated(e)
	})
}

// Deactivate deactivates the user
func (a *UserAggregate) Deactivate() error {
	if a.Status == domain.UserStatusInactive {
		return nil // Already inactive
	}

	event := UserDeactivatedEvent{
		UserID:        a.ID,
		DeactivatedAt: time.Now().UTC(),
	}

	return a.RaiseEvent(eventstore.EventTypeUserDeactivated, event, func(e *eventstore.EventEnvelope) error {
		return a.applyUserDeactivated(e)
	})
}

// Delete soft deletes the user
func (a *UserAggregate) Delete() error {
	if a.DeletedAt != nil {
		return ErrUserAlreadyDeleted
	}

	event := UserDeletedEvent{
		UserID:    a.ID,
		DeletedAt: time.Now().UTC(),
	}

	return a.RaiseEvent(eventstore.EventTypeUserDeleted, event, func(e *eventstore.EventEnvelope) error {
		return a.applyUserDeleted(e)
	})
}

// LoadFromHistory reconstructs the aggregate from events
func (a *UserAggregate) LoadFromHistory(events []*eventstore.EventEnvelope) error {
	return a.BaseAggregate.LoadFromHistory(events, func(e *eventstore.EventEnvelope) error {
		switch e.EventType {
		case eventstore.EventTypeUserCreated:
			return a.applyUserCreated(e)
		case eventstore.EventTypeUserUpdated:
			// UserUpdated can be multiple event types, check the data
			return a.applyUserUpdatedEvent(e)
		case eventstore.EventTypeUserActivated:
			return a.applyUserActivated(e)
		case eventstore.EventTypeUserDeactivated:
			return a.applyUserDeactivated(e)
		case eventstore.EventTypeUserDeleted:
			return a.applyUserDeleted(e)
		default:
			return fmt.Errorf("unknown event type: %s", e.EventType)
		}
	})
}

// Event application methods
func (a *UserAggregate) applyUserCreated(e *eventstore.EventEnvelope) error {
	var event UserCreatedEvent
	if err := e.UnmarshalData(&event); err != nil {
		return err
	}

	a.Email = event.Email
	a.Username = event.Username
	a.PasswordHash = event.PasswordHash
	a.FirstName = event.FirstName
	a.LastName = event.LastName
	a.AvatarURL = event.AvatarURL
	a.Status = event.Status
	a.CreatedAt = event.CreatedAt
	a.UpdatedAt = event.CreatedAt

	return nil
}

func (a *UserAggregate) applyUserUpdatedEvent(e *eventstore.EventEnvelope) error {
	// Try to unmarshal as different event types
	var emailChanged UserEmailChangedEvent
	if err := e.UnmarshalData(&emailChanged); err == nil && emailChanged.NewEmail != "" {
		return a.applyEmailChanged(e)
	}

	var passwordChanged UserPasswordChangedEvent
	if err := e.UnmarshalData(&passwordChanged); err == nil && passwordChanged.NewPasswordHash != "" {
		return a.applyPasswordChanged(e)
	}

	var profileUpdated UserProfileUpdatedEvent
	if err := e.UnmarshalData(&profileUpdated); err == nil {
		return a.applyProfileUpdated(e)
	}

	return fmt.Errorf("unknown user updated event format")
}

func (a *UserAggregate) applyEmailChanged(e *eventstore.EventEnvelope) error {
	var event UserEmailChangedEvent
	if err := e.UnmarshalData(&event); err != nil {
		return err
	}

	a.Email = event.NewEmail
	a.UpdatedAt = event.UpdatedAt

	return nil
}

func (a *UserAggregate) applyPasswordChanged(e *eventstore.EventEnvelope) error {
	var event UserPasswordChangedEvent
	if err := e.UnmarshalData(&event); err != nil {
		return err
	}

	a.PasswordHash = event.NewPasswordHash
	a.UpdatedAt = event.UpdatedAt

	return nil
}

func (a *UserAggregate) applyProfileUpdated(e *eventstore.EventEnvelope) error {
	var event UserProfileUpdatedEvent
	if err := e.UnmarshalData(&event); err != nil {
		return err
	}

	if event.FirstName != "" {
		a.FirstName = event.FirstName
	}
	if event.LastName != "" {
		a.LastName = event.LastName
	}
	if event.AvatarURL != "" {
		a.AvatarURL = event.AvatarURL
	}
	a.UpdatedAt = event.UpdatedAt

	return nil
}

func (a *UserAggregate) applyUserActivated(e *eventstore.EventEnvelope) error {
	var event UserActivatedEvent
	if err := e.UnmarshalData(&event); err != nil {
		return err
	}

	a.Status = domain.UserStatusActive
	a.UpdatedAt = event.ActivatedAt

	return nil
}

func (a *UserAggregate) applyUserDeactivated(e *eventstore.EventEnvelope) error {
	var event UserDeactivatedEvent
	if err := e.UnmarshalData(&event); err != nil {
		return err
	}

	a.Status = domain.UserStatusInactive
	a.UpdatedAt = event.DeactivatedAt

	return nil
}

func (a *UserAggregate) applyUserDeleted(e *eventstore.EventEnvelope) error {
	var event UserDeletedEvent
	if err := e.UnmarshalData(&event); err != nil {
		return err
	}

	a.DeletedAt = &event.DeletedAt
	a.UpdatedAt = event.DeletedAt

	return nil
}
