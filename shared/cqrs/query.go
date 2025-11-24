package cqrs

import (
	"context"
	"errors"
)

var (
	// ErrQueryHandlerNotFound is returned when no handler is registered for a query
	ErrQueryHandlerNotFound = errors.New("query handler not found")

	// ErrQueryValidation is returned when query validation fails
	ErrQueryValidation = errors.New("query validation failed")
)

// Query represents a query in the CQRS pattern
// Queries are read-only operations that return data
type Query interface {
	// QueryName returns the name of the query
	QueryName() string

	// Validate validates the query
	Validate() error
}

// QueryHandler handles a specific type of query
type QueryHandler interface {
	// Handle processes the query and returns the result
	Handle(ctx context.Context, query Query) (interface{}, error)
}

// QueryBus dispatches queries to their handlers
type QueryBus struct {
	handlers map[string]QueryHandler
}

// NewQueryBus creates a new query bus
func NewQueryBus() *QueryBus {
	return &QueryBus{
		handlers: make(map[string]QueryHandler),
	}
}

// RegisterHandler registers a query handler
func (b *QueryBus) RegisterHandler(queryName string, handler QueryHandler) {
	b.handlers[queryName] = handler
}

// Dispatch dispatches a query to its handler
func (b *QueryBus) Dispatch(ctx context.Context, query Query) (interface{}, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		return nil, ErrQueryValidation
	}

	// Get handler
	handler, ok := b.handlers[query.QueryName()]
	if !ok {
		return nil, ErrQueryHandlerNotFound
	}

	// Execute handler
	return handler.Handle(ctx, query)
}

// BaseQuery provides common functionality for queries
type BaseQuery struct{}

// Validate provides default validation (always passes)
func (q *BaseQuery) Validate() error {
	return nil
}
