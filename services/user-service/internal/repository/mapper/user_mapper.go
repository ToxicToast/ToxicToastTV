package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/user-service/internal/domain"
	"toxictoast/services/user-service/internal/repository/entity"
)

// UserToEntity converts domain model to database entity
func UserToEntity(user *domain.User) *entity.UserEntity {
	if user == nil {
		return nil
	}

	e := &entity.UserEntity{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		AvatarURL:    user.AvatarURL,
		Status:       string(user.Status),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		LastLogin:    user.LastLogin,
	}

	// Convert DeletedAt
	if user.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *user.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// UserToDomain converts database entity to domain model
func UserToDomain(e *entity.UserEntity) *domain.User {
	if e == nil {
		return nil
	}

	user := &domain.User{
		ID:           e.ID,
		Email:        e.Email,
		Username:     e.Username,
		PasswordHash: e.PasswordHash,
		FirstName:    e.FirstName,
		LastName:     e.LastName,
		AvatarURL:    e.AvatarURL,
		Status:       domain.UserStatus(e.Status),
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		LastLogin:    e.LastLogin,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		user.DeletedAt = &deletedAt
	}

	return user
}

// UsersToDomain converts slice of entities to domain models
func UsersToDomain(entities []*entity.UserEntity) []*domain.User {
	users := make([]*domain.User, 0, len(entities))
	for _, e := range entities {
		users = append(users, UserToDomain(e))
	}
	return users
}
