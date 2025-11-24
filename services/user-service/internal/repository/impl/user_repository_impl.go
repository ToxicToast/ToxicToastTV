package impl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/user-service/internal/domain"
	"toxictoast/services/user-service/internal/repository/entity"
	"toxictoast/services/user-service/internal/repository/interfaces"
	"toxictoast/services/user-service/internal/repository/mapper"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	e := mapper.UserToEntity(user)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var e entity.UserEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.UserToDomain(&e), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var e entity.UserEntity
	err := r.db.WithContext(ctx).First(&e, "email = ?", email).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.UserToDomain(&e), nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var e entity.UserEntity
	err := r.db.WithContext(ctx).First(&e, "username = ?", username).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.UserToDomain(&e), nil
}

func (r *userRepository) List(ctx context.Context, offset, limit int, status *domain.UserStatus, search *string, sortBy, sortOrder string) ([]*domain.User, int64, error) {
	var entities []*entity.UserEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.UserEntity{})

	// Filter by status
	if status != nil {
		query = query.Where("status = ?", string(*status))
	}

	// Search
	if search != nil && *search != "" {
		searchPattern := "%" + *search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	orderBy := "created_at"
	if sortBy != "" {
		switch sortBy {
		case "created_at", "updated_at", "email", "username":
			orderBy = sortBy
		}
	}

	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}

	if err := query.
		Offset(offset).
		Limit(limit).
		Order(orderBy + " " + order).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.UsersToDomain(entities), total, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	e := mapper.UserToEntity(user)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&entity.UserEntity{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.UserEntity{}).
		Where("id = ?", userID).
		Update("last_login", now).Error
}

func (r *userRepository) UpdateStatus(ctx context.Context, userID string, status domain.UserStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.UserEntity{}).
		Where("id = ?", userID).
		Update("status", string(status)).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.UserEntity{}, "id = ?", id).Error
}

func (r *userRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.UserEntity{}, "id = ?", id).Error
}
