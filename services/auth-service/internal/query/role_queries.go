package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// GetRoleQuery retrieves a role by ID
type GetRoleQuery struct {
	cqrs.BaseQuery
	RoleID string `json:"role_id"`
}

func (q *GetRoleQuery) QueryName() string {
	return "get_role"
}

func (q *GetRoleQuery) Validate() error {
	if q.RoleID == "" {
		return errors.New("role_id is required")
	}
	return nil
}

// ListRolesQuery lists all roles with pagination
type ListRolesQuery struct {
	cqrs.BaseQuery
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (q *ListRolesQuery) QueryName() string {
	return "list_roles"
}

func (q *ListRolesQuery) Validate() error {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}
	return nil
}

// ListRolesResult contains the result of a list roles query
type ListRolesResult struct {
	Roles []*domain.Role
	Total int64
}

// Query Handlers

// GetRoleHandler handles role retrieval by ID
type GetRoleHandler struct {
	roleRepo interfaces.RoleRepository
}

func NewGetRoleHandler(roleRepo interfaces.RoleRepository) *GetRoleHandler {
	return &GetRoleHandler{roleRepo: roleRepo}
}

func (h *GetRoleHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetRoleQuery)

	role, err := h.roleRepo.GetByID(ctx, q.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}

	return role, nil
}

// ListRolesHandler handles role listing
type ListRolesHandler struct {
	roleRepo interfaces.RoleRepository
}

func NewListRolesHandler(roleRepo interfaces.RoleRepository) *ListRolesHandler {
	return &ListRolesHandler{roleRepo: roleRepo}
}

func (h *ListRolesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListRolesQuery)

	offset := (q.Page - 1) * q.PageSize

	roles, total, err := h.roleRepo.List(ctx, offset, q.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return &ListRolesResult{
		Roles: roles,
		Total: total,
	}, nil
}
