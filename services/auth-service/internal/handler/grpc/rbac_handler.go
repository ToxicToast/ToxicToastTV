package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	authpb "toxictoast/services/auth-service/api/proto"
	"toxictoast/services/auth-service/internal/command"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/query"
)

// AssignRole assigns a role to a user
func (h *AuthHandler) AssignRole(ctx context.Context, req *authpb.AssignRoleRequest) (*authpb.AssignRoleResponse, error) {
	cmd := &command.AssignRoleCommand{
		BaseCommand: cqrs.BaseCommand{},
		UserID:      req.UserId,
		RoleID:      req.RoleId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign role: %v", err)
	}

	return &authpb.AssignRoleResponse{
		Success: true,
		Message: "Role assigned successfully",
	}, nil
}

// RevokeRole revokes a role from a user
func (h *AuthHandler) RevokeRole(ctx context.Context, req *authpb.RevokeRoleRequest) (*authpb.RevokeRoleResponse, error) {
	cmd := &command.RevokeRoleCommand{
		BaseCommand: cqrs.BaseCommand{},
		UserID:      req.UserId,
		RoleID:      req.RoleId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke role: %v", err)
	}

	return &authpb.RevokeRoleResponse{
		Success: true,
		Message: "Role revoked successfully",
	}, nil
}

// AssignPermission assigns a permission to a role
func (h *AuthHandler) AssignPermission(ctx context.Context, req *authpb.AssignPermissionRequest) (*authpb.AssignPermissionResponse, error) {
	cmd := &command.AssignPermissionCommand{
		BaseCommand:  cqrs.BaseCommand{},
		RoleID:       req.RoleId,
		PermissionID: req.PermissionId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign permission: %v", err)
	}

	return &authpb.AssignPermissionResponse{
		Success: true,
		Message: "Permission assigned successfully",
	}, nil
}

// RevokePermission revokes a permission from a role
func (h *AuthHandler) RevokePermission(ctx context.Context, req *authpb.RevokePermissionRequest) (*authpb.RevokePermissionResponse, error) {
	cmd := &command.RevokePermissionCommand{
		BaseCommand:  cqrs.BaseCommand{},
		RoleID:       req.RoleId,
		PermissionID: req.PermissionId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke permission: %v", err)
	}

	return &authpb.RevokePermissionResponse{
		Success: true,
		Message: "Permission revoked successfully",
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (h *AuthHandler) CheckPermission(ctx context.Context, req *authpb.CheckPermissionRequest) (*authpb.CheckPermissionResponse, error) {
	qry := &query.CheckPermissionQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.UserId,
		Resource:  req.Resource,
		Action:    req.Action,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check permission: %v", err)
	}

	checkResult := result.(*query.CheckPermissionResult)

	return &authpb.CheckPermissionResponse{
		Allowed: checkResult.HasPermission,
	}, nil
}

// ListUserRoles lists all roles for a user
func (h *AuthHandler) ListUserRoles(ctx context.Context, req *authpb.ListUserRolesRequest) (*authpb.ListUserRolesResponse, error) {
	qry := &query.GetUserRolesQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.UserId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user roles: %v", err)
	}

	roles := result.([]*domain.Role)

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
	qry := &query.GetUserPermissionsQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.UserId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user permissions: %v", err)
	}

	permissions := result.([]*domain.Permission)

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
	qry := &query.GetRolePermissionsQuery{
		BaseQuery: cqrs.BaseQuery{},
		RoleID:    req.RoleId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get role permissions: %v", err)
	}

	permissions := result.([]*domain.Permission)

	protoPermissions := make([]*authpb.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, domainPermissionToProto(permission))
	}

	return &authpb.ListRolePermissionsResponse{
		Permissions: protoPermissions,
	}, nil
}
