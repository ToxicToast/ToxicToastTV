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

// CreateRole creates a new role
func (h *AuthHandler) CreateRole(ctx context.Context, req *authpb.CreateRoleRequest) (*authpb.RoleResponse, error) {
	cmd := &command.CreateRoleCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create role: %v", err)
	}

	// Return minimal response (role was created successfully)
	return &authpb.RoleResponse{
		Role: &authpb.Role{
			Name:        req.Name,
			Description: req.Description,
		},
	}, nil
}

// GetRole retrieves a role by ID
func (h *AuthHandler) GetRole(ctx context.Context, req *authpb.GetRoleRequest) (*authpb.RoleResponse, error) {
	qry := &query.GetRoleQuery{
		BaseQuery: cqrs.BaseQuery{},
		RoleID:    req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
	}

	role := result.(*domain.Role)

	return &authpb.RoleResponse{
		Role: domainRoleToProto(role),
	}, nil
}

// UpdateRole updates an existing role
func (h *AuthHandler) UpdateRole(ctx context.Context, req *authpb.UpdateRoleRequest) (*authpb.RoleResponse, error) {
	cmd := &command.UpdateRoleCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}

	// Query the updated role
	qry := &query.GetRoleQuery{
		BaseQuery: cqrs.BaseQuery{},
		RoleID:    req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
	}

	role := result.(*domain.Role)

	return &authpb.RoleResponse{
		Role: domainRoleToProto(role),
	}, nil
}

// DeleteRole deletes a role
func (h *AuthHandler) DeleteRole(ctx context.Context, req *authpb.DeleteRoleRequest) (*authpb.DeleteResponse, error) {
	cmd := &command.DeleteRoleCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}

	return &authpb.DeleteResponse{
		Success: true,
		Message: "Role deleted successfully",
	}, nil
}

// ListRoles retrieves all roles with pagination
func (h *AuthHandler) ListRoles(ctx context.Context, req *authpb.ListRolesRequest) (*authpb.ListRolesResponse, error) {
	qry := &query.ListRolesQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	listResult := result.(*query.ListRolesResult)

	protoRoles := make([]*authpb.Role, 0, len(listResult.Roles))
	for _, role := range listResult.Roles {
		protoRoles = append(protoRoles, domainRoleToProto(role))
	}

	return &authpb.ListRolesResponse{
		Roles: protoRoles,
		Total: int32(listResult.Total),
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
