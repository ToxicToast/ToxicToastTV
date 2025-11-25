package grpc

import (
	"context"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userpb "toxictoast/services/user-service/api/proto"
	"toxictoast/services/user-service/internal/command"
	"toxictoast/services/user-service/internal/domain"
	"toxictoast/services/user-service/internal/projection"
	"toxictoast/services/user-service/internal/query"
)

// UserHandler implements the gRPC user service
type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	commandBus    *cqrs.CommandBus
	queryBus      *cqrs.QueryBus
	readModelRepo *projection.UserReadModelRepository
}

// NewUserHandler creates a new user handler with CQRS support
func NewUserHandler(
	commandBus *cqrs.CommandBus,
	queryBus *cqrs.QueryBus,
	readModelRepo *projection.UserReadModelRepository,
) *UserHandler {
	return &UserHandler{
		commandBus:    commandBus,
		queryBus:      queryBus,
		readModelRepo: readModelRepo,
	}
}

// CreateUser creates a new user using CQRS commands
func (h *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	// Create command
	cmd := &command.CreateUserCommand{
		BaseCommand: cqrs.BaseCommand{},
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		AvatarURL:   req.AvatarUrl,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Query read model for the created user
	qry := &query.GetUserByEmailQuery{
		BaseQuery: cqrs.BaseQuery{},
		Email:     req.Email,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created user: %v", err)
	}

	readModel := result.(*projection.UserReadModel)
	user := readModelToDomain(readModel)

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// GetUser retrieves a user by ID using CQRS queries
func (h *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	qry := &query.GetUserByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query user: %v", err)
	}

	readModel := result.(*projection.UserReadModel)
	if readModel == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	user := readModelToDomain(readModel)
	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// GetUserByEmail retrieves a user by email using CQRS queries
func (h *UserHandler) GetUserByEmail(ctx context.Context, req *userpb.GetUserByEmailRequest) (*userpb.UserResponse, error) {
	qry := &query.GetUserByEmailQuery{
		BaseQuery: cqrs.BaseQuery{},
		Email:     req.Email,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query user: %v", err)
	}

	readModel := result.(*projection.UserReadModel)
	if readModel == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	user := readModelToDomain(readModel)
	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// GetUserByUsername retrieves a user by username using CQRS queries
func (h *UserHandler) GetUserByUsername(ctx context.Context, req *userpb.GetUserByUsernameRequest) (*userpb.UserResponse, error) {
	qry := &query.GetUserByUsernameQuery{
		BaseQuery: cqrs.BaseQuery{},
		Username:  req.Username,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query user: %v", err)
	}

	readModel := result.(*projection.UserReadModel)
	if readModel == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	user := readModelToDomain(readModel)
	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// UpdateUser updates an existing user using CQRS commands
func (h *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserResponse, error) {
	// Get current user to check for email change
	currentUser, err := h.readModelRepo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	if currentUser == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Handle email change
	if req.Email != nil && *req.Email != currentUser.Email {
		emailCmd := &command.ChangeEmailCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
			NewEmail:    *req.Email,
		}
		if err := h.commandBus.Dispatch(ctx, emailCmd); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to change email: %v", err)
		}
	}

	// Handle profile update (firstName, lastName, avatarUrl)
	if req.FirstName != nil || req.LastName != nil || req.AvatarUrl != nil {
		profileCmd := &command.UpdateProfileCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			AvatarURL:   req.AvatarUrl,
		}

		if err := h.commandBus.Dispatch(ctx, profileCmd); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
		}
	}

	// Query updated user
	qry := &query.GetUserByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query updated user: %v", err)
	}

	readModel := result.(*projection.UserReadModel)
	user := readModelToDomain(readModel)

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// DeleteUser deletes a user using CQRS commands
func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteResponse, error) {
	cmd := &command.DeleteUserCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &userpb.DeleteResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// ListUsers retrieves users with pagination using CQRS queries
func (h *UserHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 10
	}
	page := int(req.Page)
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize

	qry := &query.ListUsersQuery{
		BaseQuery: cqrs.BaseQuery{},
		Limit:     pageSize,
		Offset:    offset,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	readModels := result.([]*projection.UserReadModel)

	protoUsers := make([]*userpb.User, 0, len(readModels))
	for _, readModel := range readModels {
		user := readModelToDomain(readModel)
		protoUsers = append(protoUsers, domainUserToProto(user))
	}

	// TODO: Add total count query for accurate pagination
	total := int32(len(readModels))
	totalPages := (total + int32(pageSize) - 1) / int32(pageSize)

	return &userpb.ListUsersResponse{
		Users:      protoUsers,
		Total:      total,
		Page:       req.Page,
		PageSize:   int32(pageSize),
		TotalPages: totalPages,
	}, nil
}

// UpdatePassword updates a user's password using CQRS commands
// Uses UpdatePasswordHashCommand because auth-service sends pre-hashed passwords
func (h *UserHandler) UpdatePassword(ctx context.Context, req *userpb.UpdatePasswordRequest) (*userpb.UpdatePasswordResponse, error) {
	cmd := &command.UpdatePasswordHashCommand{
		BaseCommand:     cqrs.BaseCommand{AggregateID: req.UserId},
		NewPasswordHash: req.PasswordHash,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update password: %v", err)
	}

	return &userpb.UpdatePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

// VerifyPassword verifies a user's password using CQRS queries
func (h *UserHandler) VerifyPassword(ctx context.Context, req *userpb.VerifyPasswordRequest) (*userpb.VerifyPasswordResponse, error) {
	// Query password hash
	qry := &query.GetUserPasswordHashQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.UserId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query password hash: %v", err)
	}

	hashResult := result.(*query.PasswordHashResult)

	// Verify password hash matches
	valid := hashResult.PasswordHash == req.PasswordHash

	return &userpb.VerifyPasswordResponse{
		Valid: valid,
	}, nil
}

// ActivateUser activates a user account using CQRS commands
func (h *UserHandler) ActivateUser(ctx context.Context, req *userpb.ActivateUserRequest) (*userpb.UserResponse, error) {
	cmd := &command.ActivateUserCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.UserId},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to activate user: %v", err)
	}

	// Query updated user
	qry := &query.GetUserByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.UserId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query activated user: %v", err)
	}

	readModel := result.(*projection.UserReadModel)
	user := readModelToDomain(readModel)

	return &userpb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// DeactivateUser deactivates a user account using CQRS commands
func (h *UserHandler) DeactivateUser(ctx context.Context, req *userpb.DeactivateUserRequest) (*userpb.UserResponse, error) {
	cmd := &command.DeactivateUserCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.UserId},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to deactivate user: %v", err)
	}

	// Query updated user
	qry := &query.GetUserByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		UserID:    req.UserId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query deactivated user: %v", err)
	}

	readModel := result.(*projection.UserReadModel)
	user := readModelToDomain(readModel)

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

// readModelToDomain converts projection.UserReadModel to domain.User
func readModelToDomain(readModel *projection.UserReadModel) *domain.User {
	return &domain.User{
		ID:        readModel.ID,
		Email:     readModel.Email,
		Username:  readModel.Username,
		FirstName: readModel.FirstName,
		LastName:  readModel.LastName,
		AvatarURL: readModel.AvatarURL,
		Status:    readModel.Status,
		CreatedAt: readModel.CreatedAt,
		UpdatedAt: readModel.LastUpdated,
		DeletedAt: readModel.DeletedAt,
	}
}
