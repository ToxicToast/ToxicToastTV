package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userpb "toxictoast/services/user-service/api/proto"
	"toxictoast/services/user-service/internal/domain"
	"toxictoast/services/user-service/internal/usecase"
)

// UserHandler implements the gRPC user service
type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	userUseCase *usecase.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	user, err := h.userUseCase.CreateUser(
		ctx,
		req.Email,
		req.Username,
		req.Password,
		req.FirstName,
		req.LastName,
		req.AvatarUrl,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	user, err := h.userUseCase.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// GetUserByEmail retrieves a user by email
func (h *UserHandler) GetUserByEmail(ctx context.Context, req *userpb.GetUserByEmailRequest) (*userpb.UserResponse, error) {
	user, err := h.userUseCase.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// GetUserByUsername retrieves a user by username
func (h *UserHandler) GetUserByUsername(ctx context.Context, req *userpb.GetUserByUsernameRequest) (*userpb.UserResponse, error) {
	user, err := h.userUseCase.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserResponse, error) {
	user, err := h.userUseCase.UpdateUser(
		ctx,
		req.Id,
		req.Email,
		req.Username,
		req.FirstName,
		req.LastName,
		req.AvatarUrl,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteResponse, error) {
	if err := h.userUseCase.DeleteUser(ctx, req.Id, req.HardDelete); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &userpb.DeleteResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// ListUsers retrieves users with pagination and filters
func (h *UserHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	var domainStatus *domain.UserStatus
	if req.Status != nil && *req.Status != userpb.UserStatus_USER_STATUS_UNSPECIFIED {
		status := protoStatusToDomain(*req.Status)
		domainStatus = &status
	}

	users, total, err := h.userUseCase.ListUsers(
		ctx,
		int(req.Page),
		int(req.PageSize),
		domainStatus,
		req.Search,
		req.SortBy,
		req.SortOrder,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	protoUsers := make([]*userpb.User, 0, len(users))
	for _, user := range users {
		protoUsers = append(protoUsers, domainUserToProto(user))
	}

	pageSize := int32(req.PageSize)
	if pageSize == 0 {
		pageSize = 10
	}
	totalPages := (int32(total) + pageSize - 1) / pageSize

	return &userpb.ListUsersResponse{
		Users:      protoUsers,
		Total:      int32(total),
		Page:       req.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdatePassword updates a user's password
func (h *UserHandler) UpdatePassword(ctx context.Context, req *userpb.UpdatePasswordRequest) (*userpb.UpdatePasswordResponse, error) {
	if err := h.userUseCase.UpdatePassword(ctx, req.UserId, req.PasswordHash); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update password: %v", err)
	}

	return &userpb.UpdatePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

// VerifyPassword verifies a user's password
func (h *UserHandler) VerifyPassword(ctx context.Context, req *userpb.VerifyPasswordRequest) (*userpb.VerifyPasswordResponse, error) {
	valid, err := h.userUseCase.VerifyPassword(ctx, req.UserId, req.PasswordHash)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify password: %v", err)
	}

	return &userpb.VerifyPasswordResponse{
		Valid: valid,
	}, nil
}

// ActivateUser activates a user account
func (h *UserHandler) ActivateUser(ctx context.Context, req *userpb.ActivateUserRequest) (*userpb.UserResponse, error) {
	user, err := h.userUseCase.ActivateUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to activate user: %v", err)
	}

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// DeactivateUser deactivates a user account
func (h *UserHandler) DeactivateUser(ctx context.Context, req *userpb.DeactivateUserRequest) (*userpb.UserResponse, error) {
	user, err := h.userUseCase.DeactivateUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to deactivate user: %v", err)
	}

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// domainUserToProto converts domain.User to userpb.User
func domainUserToProto(user *domain.User) *userpb.User {
	protoUser := &userpb.User{
		Id:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: user.AvatarURL,
		Status:    domainStatusToProto(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	if user.LastLogin != nil {
		protoUser.LastLogin = timestamppb.New(*user.LastLogin)
	}

	return protoUser
}

// domainStatusToProto converts domain.UserStatus to userpb.UserStatus
func domainStatusToProto(status domain.UserStatus) userpb.UserStatus {
	switch status {
	case domain.UserStatusActive:
		return userpb.UserStatus_USER_STATUS_ACTIVE
	case domain.UserStatusInactive:
		return userpb.UserStatus_USER_STATUS_INACTIVE
	case domain.UserStatusSuspended:
		return userpb.UserStatus_USER_STATUS_SUSPENDED
	case domain.UserStatusDeleted:
		return userpb.UserStatus_USER_STATUS_DELETED
	default:
		return userpb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

// protoStatusToDomain converts userpb.UserStatus to domain.UserStatus
func protoStatusToDomain(status userpb.UserStatus) domain.UserStatus {
	switch status {
	case userpb.UserStatus_USER_STATUS_ACTIVE:
		return domain.UserStatusActive
	case userpb.UserStatus_USER_STATUS_INACTIVE:
		return domain.UserStatusInactive
	case userpb.UserStatus_USER_STATUS_SUSPENDED:
		return domain.UserStatusSuspended
	case userpb.UserStatus_USER_STATUS_DELETED:
		return domain.UserStatusDeleted
	default:
		return domain.UserStatusActive
	}
}
