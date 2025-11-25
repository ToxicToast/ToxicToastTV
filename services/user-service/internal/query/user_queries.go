package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/eventstore"
	"toxictoast/services/user-service/internal/aggregate"
	"toxictoast/services/user-service/internal/projection"
)

// GetUserByIDQuery retrieves a user by ID
type GetUserByIDQuery struct {
	cqrs.BaseQuery
	UserID string `json:"user_id"`
}

func (q *GetUserByIDQuery) QueryName() string {
	return "get_user_by_id"
}

func (q *GetUserByIDQuery) Validate() error {
	if q.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// GetUserByEmailQuery retrieves a user by email
type GetUserByEmailQuery struct {
	cqrs.BaseQuery
	Email string `json:"email"`
}

func (q *GetUserByEmailQuery) QueryName() string {
	return "get_user_by_email"
}

func (q *GetUserByEmailQuery) Validate() error {
	if q.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

// GetUserByUsernameQuery retrieves a user by username
type GetUserByUsernameQuery struct {
	cqrs.BaseQuery
	Username string `json:"username"`
}

func (q *GetUserByUsernameQuery) QueryName() string {
	return "get_user_by_username"
}

func (q *GetUserByUsernameQuery) Validate() error {
	if q.Username == "" {
		return errors.New("username is required")
	}
	return nil
}

// ListUsersQuery lists all users with pagination
type ListUsersQuery struct {
	cqrs.BaseQuery
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (q *ListUsersQuery) QueryName() string {
	return "list_users"
}

func (q *ListUsersQuery) Validate() error {
	if q.Limit <= 0 {
		q.Limit = 10
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return nil
}

// Query Handlers

// GetUserByIDHandler handles user retrieval by ID
type GetUserByIDHandler struct {
	repo *projection.UserReadModelRepository
}

func NewGetUserByIDHandler(repo *projection.UserReadModelRepository) *GetUserByIDHandler {
	return &GetUserByIDHandler{repo: repo}
}

func (h *GetUserByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetUserByIDQuery)
	return h.repo.FindByID(ctx, q.UserID)
}

// GetUserByEmailHandler handles user retrieval by email
type GetUserByEmailHandler struct {
	repo *projection.UserReadModelRepository
}

func NewGetUserByEmailHandler(repo *projection.UserReadModelRepository) *GetUserByEmailHandler {
	return &GetUserByEmailHandler{repo: repo}
}

func (h *GetUserByEmailHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetUserByEmailQuery)
	return h.repo.FindByEmail(ctx, q.Email)
}

// GetUserByUsernameHandler handles user retrieval by username
type GetUserByUsernameHandler struct {
	repo *projection.UserReadModelRepository
}

func NewGetUserByUsernameHandler(repo *projection.UserReadModelRepository) *GetUserByUsernameHandler {
	return &GetUserByUsernameHandler{repo: repo}
}

func (h *GetUserByUsernameHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetUserByUsernameQuery)
	return h.repo.FindByUsername(ctx, q.Username)
}

// ListUsersHandler handles user listing
type ListUsersHandler struct {
	repo *projection.UserReadModelRepository
}

func NewListUsersHandler(repo *projection.UserReadModelRepository) *ListUsersHandler {
	return &ListUsersHandler{repo: repo}
}

func (h *ListUsersHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListUsersQuery)
	return h.repo.FindAll(ctx, q.Limit, q.Offset)
}

// GetUserPasswordHashQuery retrieves a user's password hash for verification
type GetUserPasswordHashQuery struct {
	cqrs.BaseQuery
	UserID string `json:"user_id"`
}

func (q *GetUserPasswordHashQuery) QueryName() string {
	return "get_user_password_hash"
}

func (q *GetUserPasswordHashQuery) Validate() error {
	if q.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// GetUserPasswordHashHandler handles password hash retrieval
type GetUserPasswordHashHandler struct {
	eventStore eventstore.EventStore
}

func NewGetUserPasswordHashHandler(eventStore eventstore.EventStore) *GetUserPasswordHashHandler {
	return &GetUserPasswordHashHandler{eventStore: eventStore}
}

// PasswordHashResult contains the password hash
type PasswordHashResult struct {
	PasswordHash string `json:"password_hash"`
}

func (h *GetUserPasswordHashHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetUserPasswordHashQuery)

	// Load user aggregate from event store to get password hash
	user := aggregate.NewUserAggregate(q.UserID)

	// Get events from event store
	events, err := h.eventStore.GetEvents(ctx, eventstore.AggregateTypeUser, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, errors.New("user not found")
	}

	// Reconstruct aggregate from events
	if err := user.LoadFromHistory(events); err != nil {
		return nil, fmt.Errorf("failed to load user from history: %w", err)
	}

	return &PasswordHashResult{
		PasswordHash: user.PasswordHash,
	}, nil
}
