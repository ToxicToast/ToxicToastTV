package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	authpb "toxictoast/services/auth-service/api/proto"
	"toxictoast/services/auth-service/internal/domain"
)

// CreatePermission creates a new permission
func (h *AuthHandler) CreatePermission(ctx context.Context, req *authpb.CreatePermissionRequest) (*authpb.PermissionResponse, error) {
	permission, err := h.permissionUseCase.CreatePermission(ctx, req.Resource, req.Action, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create permission: %v", err)
	}

	return &authpb.PermissionResponse{
		Permission: domainPermissionToProto(permission),
	}, nil
}

// GetPermission retrieves a permission by ID
func (h *AuthHandler) GetPermission(ctx context.Context, req *authpb.GetPermissionRequest) (*authpb.PermissionResponse, error) {
	permission, err := h.permissionUseCase.GetPermission(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "permission not found: %v", err)
	}

	return &authpb.PermissionResponse{
		Permission: domainPermissionToProto(permission),
	}, nil
}

// UpdatePermission updates an existing permission
func (h *AuthHandler) UpdatePermission(ctx context.Context, req *authpb.UpdatePermissionRequest) (*authpb.PermissionResponse, error) {
	permission, err := h.permissionUseCase.UpdatePermission(ctx, req.Id, req.Resource, req.Action, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update permission: %v", err)
	}

	return &authpb.PermissionResponse{
		Permission: domainPermissionToProto(permission),
	}, nil
}

// DeletePermission deletes a permission
func (h *AuthHandler) DeletePermission(ctx context.Context, req *authpb.DeletePermissionRequest) (*authpb.DeleteResponse, error) {
	if err := h.permissionUseCase.DeletePermission(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete permission: %v", err)
	}

	return &authpb.DeleteResponse{
		Success: true,
		Message: "Permission deleted successfully",
	}, nil
}

// ListPermissions retrieves all permissions with pagination
func (h *AuthHandler) ListPermissions(ctx context.Context, req *authpb.ListPermissionsRequest) (*authpb.ListPermissionsResponse, error) {
	permissions, total, err := h.permissionUseCase.ListPermissions(ctx, int(req.Page), int(req.PageSize), req.Resource)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions: %v", err)
	}

	protoPermissions := make([]*authpb.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, domainPermissionToProto(permission))
	}

	return &authpb.ListPermissionsResponse{
		Permissions: protoPermissions,
		Total:       int32(total),
	}, nil
}

// domainPermissionToProto converts domain.Permission to authpb.Permission
func domainPermissionToProto(permission *domain.Permission) *authpb.Permission {
	return &authpb.Permission{
		Id:          permission.ID,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		CreatedAt:   timestamppb.New(permission.CreatedAt),
		UpdatedAt:   timestamppb.New(permission.UpdatedAt),
	}
}
