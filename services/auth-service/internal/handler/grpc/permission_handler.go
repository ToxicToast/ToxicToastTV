package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	authpb "toxictoast/services/auth-service/api/proto"
	"toxictoast/services/auth-service/internal/command"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/query"
)

// CreatePermission creates a new permission
func (h *AuthHandler) CreatePermission(ctx context.Context, req *authpb.CreatePermissionRequest) (*authpb.PermissionResponse, error) {
	cmd := &command.CreatePermissionCommand{
		BaseCommand: cqrs.BaseCommand{},
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create permission: %v", err)
	}

	return &authpb.PermissionResponse{
		Permission: &authpb.Permission{
			Resource:    req.Resource,
			Action:      req.Action,
			Description: req.Description,
		},
	}, nil
}

// GetPermission retrieves a permission by ID
func (h *AuthHandler) GetPermission(ctx context.Context, req *authpb.GetPermissionRequest) (*authpb.PermissionResponse, error) {
	qry := &query.GetPermissionQuery{
		BaseQuery:    cqrs.BaseQuery{},
		PermissionID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "permission not found: %v", err)
	}

	permission := result.(*domain.Permission)

	return &authpb.PermissionResponse{
		Permission: domainPermissionToProto(permission),
	}, nil
}

// UpdatePermission updates an existing permission
func (h *AuthHandler) UpdatePermission(ctx context.Context, req *authpb.UpdatePermissionRequest) (*authpb.PermissionResponse, error) {
	cmd := &command.UpdatePermissionCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update permission: %v", err)
	}

	// Query the updated permission
	qry := &query.GetPermissionQuery{
		BaseQuery:    cqrs.BaseQuery{},
		PermissionID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "permission not found: %v", err)
	}

	permission := result.(*domain.Permission)

	return &authpb.PermissionResponse{
		Permission: domainPermissionToProto(permission),
	}, nil
}

// DeletePermission deletes a permission
func (h *AuthHandler) DeletePermission(ctx context.Context, req *authpb.DeletePermissionRequest) (*authpb.DeleteResponse, error) {
	cmd := &command.DeletePermissionCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete permission: %v", err)
	}

	return &authpb.DeleteResponse{
		Success: true,
		Message: "Permission deleted successfully",
	}, nil
}

// ListPermissions retrieves all permissions with pagination
func (h *AuthHandler) ListPermissions(ctx context.Context, req *authpb.ListPermissionsRequest) (*authpb.ListPermissionsResponse, error) {
	qry := &query.ListPermissionsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
		Resource:  req.Resource,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions: %v", err)
	}

	listResult := result.(*query.ListPermissionsResult)

	protoPermissions := make([]*authpb.Permission, 0, len(listResult.Permissions))
	for _, perm := range listResult.Permissions {
		protoPermissions = append(protoPermissions, domainPermissionToProto(perm))
	}

	return &authpb.ListPermissionsResponse{
		Permissions: protoPermissions,
		Total:       int32(listResult.Total),
	}, nil
}

// domainPermissionToProto converts domain.Permission to authpb.Permission
func domainPermissionToProto(perm *domain.Permission) *authpb.Permission {
	return &authpb.Permission{
		Id:          perm.ID,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: perm.Description,
		CreatedAt:   timestamppb.New(perm.CreatedAt),
		UpdatedAt:   timestamppb.New(perm.UpdatedAt),
	}
}
