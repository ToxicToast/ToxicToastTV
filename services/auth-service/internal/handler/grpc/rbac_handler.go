package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "toxictoast/services/auth-service/api/proto"
)

// AssignRole assigns a role to a user
func (h *AuthHandler) AssignRole(ctx context.Context, req *authpb.AssignRoleRequest) (*authpb.AssignRoleResponse, error) {
	if err := h.rbacUseCase.AssignRole(ctx, req.UserId, req.RoleId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign role: %v", err)
	}

	return &authpb.AssignRoleResponse{
		Success: true,
		Message: "Role assigned successfully",
	}, nil
}

// RevokeRole revokes a role from a user
func (h *AuthHandler) RevokeRole(ctx context.Context, req *authpb.RevokeRoleRequest) (*authpb.RevokeRoleResponse, error) {
	if err := h.rbacUseCase.RevokeRole(ctx, req.UserId, req.RoleId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke role: %v", err)
	}

	return &authpb.RevokeRoleResponse{
		Success: true,
		Message: "Role revoked successfully",
	}, nil
}

// AssignPermission assigns a permission to a role
func (h *AuthHandler) AssignPermission(ctx context.Context, req *authpb.AssignPermissionRequest) (*authpb.AssignPermissionResponse, error) {
	if err := h.rbacUseCase.AssignPermission(ctx, req.RoleId, req.PermissionId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign permission: %v", err)
	}

	return &authpb.AssignPermissionResponse{
		Success: true,
		Message: "Permission assigned successfully",
	}, nil
}

// RevokePermission revokes a permission from a role
func (h *AuthHandler) RevokePermission(ctx context.Context, req *authpb.RevokePermissionRequest) (*authpb.RevokePermissionResponse, error) {
	if err := h.rbacUseCase.RevokePermission(ctx, req.RoleId, req.PermissionId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke permission: %v", err)
	}

	return &authpb.RevokePermissionResponse{
		Success: true,
		Message: "Permission revoked successfully",
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (h *AuthHandler) CheckPermission(ctx context.Context, req *authpb.CheckPermissionRequest) (*authpb.CheckPermissionResponse, error) {
	allowed, err := h.rbacUseCase.CheckPermission(ctx, req.UserId, req.Resource, req.Action)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check permission: %v", err)
	}

	return &authpb.CheckPermissionResponse{
		Allowed: allowed,
	}, nil
}

// ListUserRoles lists all roles for a user
func (h *AuthHandler) ListUserRoles(ctx context.Context, req *authpb.ListUserRolesRequest) (*authpb.ListUserRolesResponse, error) {
	roles, err := h.rbacUseCase.GetUserRoles(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user roles: %v", err)
	}

	protoRoles := make([]*authpb.Role, 0, len(roles))
	for _, role := range roles {
		protoRoles = append(protoRoles, domainRoleToProto(role))
	}

	return &authpb.ListUserRolesResponse{
		Roles: protoRoles,
	}, nil
}

// ListUserPermissions lists all permissions for a user
func (h *AuthHandler) ListUserPermissions(ctx context.Context, req *authpb.ListUserPermissionsRequest) (*authpb.ListUserPermissionsResponse, error) {
	permissions, err := h.rbacUseCase.GetUserPermissions(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user permissions: %v", err)
	}

	protoPermissions := make([]*authpb.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, domainPermissionToProto(permission))
	}

	return &authpb.ListUserPermissionsResponse{
		Permissions: protoPermissions,
	}, nil
}

// ListRolePermissions lists all permissions for a role
func (h *AuthHandler) ListRolePermissions(ctx context.Context, req *authpb.ListRolePermissionsRequest) (*authpb.ListRolePermissionsResponse, error) {
	permissions, err := h.rbacUseCase.GetRolePermissions(ctx, req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get role permissions: %v", err)
	}

	protoPermissions := make([]*authpb.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, domainPermissionToProto(permission))
	}

	return &authpb.ListRolePermissionsResponse{
		Permissions: protoPermissions,
	}, nil
}
