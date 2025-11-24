package projection

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/eventstore"
	"toxictoast/services/user-service/internal/aggregate"
	"toxictoast/services/user-service/internal/domain"
)

// UserReadModel represents the read-optimized user view
type UserReadModel struct {
	*cqrs.BaseReadModel
	Email      string            `json:"email" db:"email"`
	Username   string            `json:"username" db:"username"`
	FirstName  string            `json:"first_name" db:"first_name"`
	LastName   string            `json:"last_name" db:"last_name"`
	AvatarURL  string            `json:"avatar_url" db:"avatar_url"`
	Status     domain.UserStatus `json:"status" db:"status"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	DeletedAt  *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserReadModelRepository manages user read models
type UserReadModelRepository struct {
	db *sql.DB
}

// NewUserReadModelRepository creates a new repository
func NewUserReadModelRepository(db *sql.DB) *UserReadModelRepository {
	return &UserReadModelRepository{db: db}
}

// CreateTable creates the read model table
func (r *UserReadModelRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS user_read_model (
		id VARCHAR(36) PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		username VARCHAR(255) UNIQUE NOT NULL,
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		avatar_url TEXT,
		status VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		last_updated TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP,
		CONSTRAINT chk_status CHECK (status IN ('active', 'inactive', 'suspended'))
	);

	CREATE INDEX IF NOT EXISTS idx_user_read_model_email ON user_read_model(email);
	CREATE INDEX IF NOT EXISTS idx_user_read_model_username ON user_read_model(username);
	CREATE INDEX IF NOT EXISTS idx_user_read_model_status ON user_read_model(status);
	CREATE INDEX IF NOT EXISTS idx_user_read_model_created_at ON user_read_model(created_at);
	`

	_, err := r.db.Exec(query)
	return err
}

// FindByID finds a user by ID
func (r *UserReadModelRepository) FindByID(ctx context.Context, id string) (*UserReadModel, error) {
	var user UserReadModel
	user.BaseReadModel = &cqrs.BaseReadModel{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, username, first_name, last_name, avatar_url, status, created_at, last_updated, deleted_at
		FROM user_read_model
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.LastUpdated,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserReadModelRepository) FindByEmail(ctx context.Context, email string) (*UserReadModel, error) {
	var user UserReadModel
	user.BaseReadModel = &cqrs.BaseReadModel{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, username, first_name, last_name, avatar_url, status, created_at, last_updated, deleted_at
		FROM user_read_model
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.LastUpdated,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByUsername finds a user by username
func (r *UserReadModelRepository) FindByUsername(ctx context.Context, username string) (*UserReadModel, error) {
	var user UserReadModel
	user.BaseReadModel = &cqrs.BaseReadModel{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, username, first_name, last_name, avatar_url, status, created_at, last_updated, deleted_at
		FROM user_read_model
		WHERE username = $1 AND deleted_at IS NULL
	`, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.LastUpdated,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindAll returns all users with pagination
func (r *UserReadModelRepository) FindAll(ctx context.Context, limit, offset int) ([]*UserReadModel, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, username, first_name, last_name, avatar_url, status, created_at, last_updated, deleted_at
		FROM user_read_model
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer rows.Close()

	var users []*UserReadModel
	for rows.Next() {
		var user UserReadModel
		user.BaseReadModel = &cqrs.BaseReadModel{}

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.AvatarURL,
			&user.Status,
			&user.CreatedAt,
			&user.LastUpdated,
			&user.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, &user)
	}

	return users, nil
}

// Save upserts a user read model
func (r *UserReadModelRepository) Save(ctx context.Context, user *UserReadModel) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_read_model (
			id, email, username, first_name, last_name, avatar_url, status, created_at, last_updated, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			avatar_url = EXCLUDED.avatar_url,
			status = EXCLUDED.status,
			last_updated = EXCLUDED.last_updated,
			deleted_at = EXCLUDED.deleted_at
	`,
		user.ID,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.Status,
		user.CreatedAt,
		user.LastUpdated,
		user.DeletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save user read model: %w", err)
	}

	return nil
}

// Delete removes a user read model
func (r *UserReadModelRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_read_model WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user read model: %w", err)
	}
	return nil
}

// UserProjector projects user events into read models
type UserProjector struct {
	repo *UserReadModelRepository
}

// NewUserProjector creates a new user projector
func NewUserProjector(repo *UserReadModelRepository) *UserProjector {
	return &UserProjector{repo: repo}
}

// GetProjectorName returns the projector name
func (p *UserProjector) GetProjectorName() string {
	return "user_projector"
}

// GetEventTypes returns the event types this projector handles
func (p *UserProjector) GetEventTypes() []string {
	return []string{
		eventstore.EventTypeUserCreated,
		eventstore.EventTypeUserUpdated,
		eventstore.EventTypeUserActivated,
		eventstore.EventTypeUserDeactivated,
		eventstore.EventTypeUserDeleted,
	}
}

// ProjectEvent projects an event into the read model
func (p *UserProjector) ProjectEvent(ctx context.Context, event *eventstore.EventEnvelope) error {
	switch event.EventType {
	case eventstore.EventTypeUserCreated:
		return p.projectUserCreated(ctx, event)
	case eventstore.EventTypeUserUpdated:
		return p.projectUserUpdated(ctx, event)
	case eventstore.EventTypeUserActivated:
		return p.projectUserActivated(ctx, event)
	case eventstore.EventTypeUserDeactivated:
		return p.projectUserDeactivated(ctx, event)
	case eventstore.EventTypeUserDeleted:
		return p.projectUserDeleted(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

func (p *UserProjector) projectUserCreated(ctx context.Context, event *eventstore.EventEnvelope) error {
	var e aggregate.UserCreatedEvent
	if err := event.UnmarshalData(&e); err != nil {
		return err
	}

	user := &UserReadModel{
		BaseReadModel: cqrs.NewBaseReadModel(event.AggregateID),
		Email:         e.Email,
		Username:      e.Username,
		FirstName:     e.FirstName,
		LastName:      e.LastName,
		AvatarURL:     e.AvatarURL,
		Status:        e.Status,
		CreatedAt:     e.CreatedAt,
	}

	return p.repo.Save(ctx, user)
}

func (p *UserProjector) projectUserUpdated(ctx context.Context, event *eventstore.EventEnvelope) error {
	// Load existing read model
	user, err := p.repo.FindByID(ctx, event.AggregateID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", event.AggregateID)
	}

	// Try to unmarshal as different update events
	var emailChanged aggregate.UserEmailChangedEvent
	if err := event.UnmarshalData(&emailChanged); err == nil && emailChanged.NewEmail != "" {
		user.Email = emailChanged.NewEmail
		user.UpdateTimestamp()
		return p.repo.Save(ctx, user)
	}

	var profileUpdated aggregate.UserProfileUpdatedEvent
	if err := event.UnmarshalData(&profileUpdated); err == nil {
		if profileUpdated.FirstName != "" {
			user.FirstName = profileUpdated.FirstName
		}
		if profileUpdated.LastName != "" {
			user.LastName = profileUpdated.LastName
		}
		if profileUpdated.AvatarURL != "" {
			user.AvatarURL = profileUpdated.AvatarURL
		}
		user.UpdateTimestamp()
		return p.repo.Save(ctx, user)
	}

	// Password changes don't affect read model
	return nil
}

func (p *UserProjector) projectUserActivated(ctx context.Context, event *eventstore.EventEnvelope) error {
	user, err := p.repo.FindByID(ctx, event.AggregateID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", event.AggregateID)
	}

	user.Status = domain.UserStatusActive
	user.UpdateTimestamp()

	return p.repo.Save(ctx, user)
}

func (p *UserProjector) projectUserDeactivated(ctx context.Context, event *eventstore.EventEnvelope) error {
	user, err := p.repo.FindByID(ctx, event.AggregateID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", event.AggregateID)
	}

	user.Status = domain.UserStatusInactive
	user.UpdateTimestamp()

	return p.repo.Save(ctx, user)
}

func (p *UserProjector) projectUserDeleted(ctx context.Context, event *eventstore.EventEnvelope) error {
	var e aggregate.UserDeletedEvent
	if err := event.UnmarshalData(&e); err != nil {
		return err
	}

	user, err := p.repo.FindByID(ctx, event.AggregateID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", event.AggregateID)
	}

	user.DeletedAt = &e.DeletedAt
	user.UpdateTimestamp()

	return p.repo.Save(ctx, user)
}
