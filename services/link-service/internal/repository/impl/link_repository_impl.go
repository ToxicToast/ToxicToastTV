package impl

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
)

type linkRepository struct {
	db *gorm.DB
}

// NewLinkRepository creates a new instance of LinkRepository
func NewLinkRepository(db *gorm.DB) repository.LinkRepository {
	return &linkRepository{db: db}
}

func (r *linkRepository) Create(ctx context.Context, link *domain.Link) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *linkRepository) GetByID(ctx context.Context, id string) (*domain.Link, error) {
	var link domain.Link
	err := r.db.WithContext(ctx).First(&link, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("link not found")
		}
		return nil, err
	}
	return &link, nil
}

func (r *linkRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.Link, error) {
	var link domain.Link
	err := r.db.WithContext(ctx).First(&link, "short_code = ?", shortCode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("link not found")
		}
		return nil, err
	}
	return &link, nil
}

func (r *linkRepository) List(ctx context.Context, filters repository.LinkFilters) ([]domain.Link, int64, error) {
	var links []domain.Link
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Link{})

	// Apply filters
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	// Filter expired links
	if !filters.IncludeExpired {
		query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())
	}

	// Search filter
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("original_url ILIKE ? OR title ILIKE ? OR description ILIKE ? OR short_code ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}

	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Pagination
	if filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	// Execute query
	if err := query.Find(&links).Error; err != nil {
		return nil, 0, err
	}

	return links, total, nil
}

func (r *linkRepository) Update(ctx context.Context, link *domain.Link) error {
	return r.db.WithContext(ctx).Save(link).Error
}

func (r *linkRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Link{}, "id = ?", id).Error
}

func (r *linkRepository) IncrementClicks(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Link{}).
		Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

func (r *linkRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Link{}).Where("short_code = ?", shortCode).Count(&count).Error
	return count > 0, err
}

func (r *linkRepository) GetStats(ctx context.Context, linkID string) (*repository.LinkStats, error) {
	stats := &repository.LinkStats{
		LinkID: linkID,
	}

	// Get total clicks from link
	var link domain.Link
	if err := r.db.WithContext(ctx).First(&link, "id = ?", linkID).Error; err != nil {
		return nil, err
	}
	stats.TotalClicks = link.ClickCount

	// Get unique IPs
	var uniqueIPs int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Click{}).
		Where("link_id = ?", linkID).
		Distinct("ip_address").
		Count(&uniqueIPs).Error; err != nil {
		return nil, err
	}
	stats.UniqueIPs = int(uniqueIPs)

	// Get clicks today
	today := time.Now().Truncate(24 * time.Hour)
	var clicksToday int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Click{}).
		Where("link_id = ? AND clicked_at >= ?", linkID, today).
		Count(&clicksToday).Error; err != nil {
		return nil, err
	}
	stats.ClicksToday = int(clicksToday)

	// Get clicks this week
	weekStart := time.Now().AddDate(0, 0, -7)
	var clicksWeek int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Click{}).
		Where("link_id = ? AND clicked_at >= ?", linkID, weekStart).
		Count(&clicksWeek).Error; err != nil {
		return nil, err
	}
	stats.ClicksWeek = int(clicksWeek)

	// Get clicks this month
	monthStart := time.Now().AddDate(0, -1, 0)
	var clicksMonth int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Click{}).
		Where("link_id = ? AND clicked_at >= ?", linkID, monthStart).
		Count(&clicksMonth).Error; err != nil {
		return nil, err
	}
	stats.ClicksMonth = int(clicksMonth)

	return stats, nil
}
