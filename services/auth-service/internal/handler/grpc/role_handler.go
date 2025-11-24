package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	authpb "toxictoast/services/auth-service/api/proto"
	"toxictoast/services/auth-service/internal/domain"
)

// CreateRole creates a new role
func (h *AuthHandler) CreateRole(ctx context.Context, req *authpb.CreateRoleRequest) (*authpb.RoleResponse, error) {
	role, err := h.roleUseCase.CreateRole(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create role: %v", err)
	}

	return &authpb.RoleResponse{
		Role: domainRoleToProto(role),
	}, nil
}

// GetRole retrieves a role by ID
func (h *AuthHandler) GetRole(ctx context.Context, req *authpb.GetRoleRequest) (*authpb.RoleResponse, error) {
	role, err := h.roleUseCase.GetRole(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
	}

	return &authpb.RoleResponse{
		Role: domainRoleToProto(role),
	}, nil
}

// UpdateRole updates an existing role
func (h *AuthHandler) UpdateRole(ctx context.Context, req *authpb.UpdateRoleRequest) (*authpb.RoleResponse, error) {
	role, err := h.roleUseCase.UpdateRole(ctx, req.Id, req.Name, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}

	return &authpb.RoleResponse{
		Role: domainRoleToProto(role),
	}, nil
}

// DeleteRole deletes a role
func (h *AuthHandler) DeleteRole(ctx context.Context, req *authpb.DeleteRoleRequest) (*authpb.DeleteResponse, error) {
	if err := h.roleUseCase.DeleteRole(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}

	return &authpb.DeleteResponse{
		Success: true,
		Message: "Role deleted successfully",
	}, nil
}

// ListRoles retrieves all roles with pagination
func (h *AuthHandler) ListRoles(ctx context.Context, req *authpb.ListRolesRequest) (*authpb.ListRolesResponse, error) {
	roles, total, err := h.roleUseCase.ListRoles(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	protoRoles := make([]*authpb.Role, 0, len(roles))
	for _, role := range roles {
		protoRoles = append(protoRoles, domainRoleToProto(role))
	}

	return &authpb.ListRolesResponse{
		Roles: protoRoles,
		Total: int32(total),
	}, nil
}

// domainRoleToProto converts domain.Role to authpb.Role
func domainRoleToProto(role *domain.Role) *authpb.Role {
	return &authpb.Role{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}
}
