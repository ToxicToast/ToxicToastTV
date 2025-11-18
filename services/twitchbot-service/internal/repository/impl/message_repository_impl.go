package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"
	"toxictoast/services/twitchbot-service/internal/repository/mapper"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository instance
func NewMessageRepository(db *gorm.DB) interfaces.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Create(mapper.MessageToEntity(message)).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	var message domain.Message
	err := r.db.WithContext(ctx).First(&message, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) List(ctx context.Context, offset, limit int, streamID, userID string, includeDeleted bool) ([]*domain.Message, int64, error) {
	var messages []*domain.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.MessageEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if streamID != "" {
		query = query.Where("stream_id = ?", streamID)
	}

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("sent_at DESC").Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) Search(ctx context.Context, query string, streamID, userID string, offset, limit int) ([]*domain.Message, int64, error) {
	var messages []*domain.Message
	var total int64

	dbQuery := r.db.WithContext(ctx).Model(&entity.MessageEntity{})

	if query != "" {
		dbQuery = dbQuery.Where("message ILIKE ?", "%"+query+"%")
	}

	if streamID != "" {
		dbQuery = dbQuery.Where("stream_id = ?", streamID)
	}

	if userID != "" {
		dbQuery = dbQuery.Where("user_id = ?", userID)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := dbQuery.Order("sent_at DESC").Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) GetStats(ctx context.Context, streamID string) (totalMessages int64, uniqueUsers int64, mostActiveUser string, mostActiveUserCount int64, err error) {
	// Total messages
	err = r.db.WithContext(ctx).Model(&entity.MessageEntity{}).Where("stream_id = ?", streamID).Count(&totalMessages).Error
	if err != nil {
		return 0, 0, "", 0, err
	}

	// Unique users
	err = r.db.WithContext(ctx).Model(&entity.MessageEntity{}).Where("stream_id = ?", streamID).Distinct("user_id").Count(&uniqueUsers).Error
	if err != nil {
		return 0, 0, "", 0, err
	}

	// Most active user
	type Result struct {
		UserID string
		Count  int64
	}
	var result Result
	err = r.db.WithContext(ctx).Model(&entity.MessageEntity{}).
		Select("user_id, COUNT(*) as count").
		Where("stream_id = ?", streamID).
		Group("user_id").
		Order("count DESC").
		Limit(1).
		Scan(&result).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, "", 0, err
	}

	return totalMessages, uniqueUsers, result.UserID, result.Count, nil
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.MessageEntity{}, "id = ?", id).Error
}

func (r *messageRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.MessageEntity{}, "id = ?", id).Error
}
