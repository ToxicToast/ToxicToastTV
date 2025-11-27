package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/jwt"
)

// ValidateTokenQuery validates a JWT token
type ValidateTokenQuery struct {
	cqrs.BaseQuery
	Token string `json:"token"`
}

func (q *ValidateTokenQuery) QueryName() string {
	return "validate_token"
}

func (q *ValidateTokenQuery) Validate() error {
	if q.Token == "" {
		return errors.New("token is required")
	}
	return nil
}

// Query Handlers

// ValidateTokenHandler handles token validation
type ValidateTokenHandler struct {
	jwtHelper *jwt.JWTHelper
}

func NewValidateTokenHandler(jwtHelper *jwt.JWTHelper) *ValidateTokenHandler {
	return &ValidateTokenHandler{jwtHelper: jwtHelper}
}

func (h *ValidateTokenHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ValidateTokenQuery)

	claims, err := h.jwtHelper.ValidateToken(q.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
