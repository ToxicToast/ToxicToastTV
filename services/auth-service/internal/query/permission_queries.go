package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// GetPermissionQuery retrieves a permission by ID
type GetPermissionQuery struct {
	cqrs.BaseQuery
	PermissionID string `json:"permission_id"`
}

func (q *GetPermissionQuery) QueryName() string {
	return "get_permission"
}

func (q *GetPermissionQuery) Validate() error {
	if q.PermissionID == "" {
		return errors.New("permission_id is required")
	}
	return nil
}

// ListPermissionsQuery lists all permissions with pagination
type ListPermissionsQuery struct {
	cqrs.BaseQuery
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	Resource *string `json:"resource,omitempty"`
}

func (q *ListPermissionsQuery) QueryName() string {
	return "list_permissions"
}

func (q *ListPermissionsQuery) Validate() error {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}
	return nil
}

// ListPermissionsResult contains the result of a list permissions query
type ListPermissionsResult struct {
	Permissions []*domain.Permission
	Total       int64
}

// Query Handlers

// GetPermissionHandler handles permission retrieval by ID
type GetPermissionHandler struct {
	permissionRepo interfaces.PermissionRepository
}

func NewGetPermissionHandler(permissionRepo interfaces.PermissionRepository) *GetPermissionHandler {
	return &GetPermissionHandler{permissionRepo: permissionRepo}
}

func (h *GetPermissionHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetPermissionQuery)

	permission, err := h.permissionRepo.GetByID(ctx, q.PermissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return nil, fmt.Errorf("permission not found")
	}

	return permission, nil
}

// ListPermissionsHandler handles permission listing
type ListPermissionsHandler struct {
	permissionRepo interfaces.PermissionRepository
}

func NewListPermissionsHandler(permissionRepo interfaces.PermissionRepository) *ListPermissionsHandler {
	return &ListPermissionsHandler{permissionRepo: permissionRepo}
}

func (h *ListPermissionsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListPermissionsQuery)

	offset := (q.Page - 1) * q.PageSize

	permissions, total, err := h.permissionRepo.List(ctx, offset, q.PageSize, q.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return &ListPermissionsResult{
		Permissions: permissions,
		Total:       total,
	}, nil
}
