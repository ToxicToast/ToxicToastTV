package impl

import (
	"context"
	"time"

	"gorm.io/gorm"
	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
	"toxictoast/services/link-service/internal/repository/entity"
	"toxictoast/services/link-service/internal/repository/mapper"
)

type clickRepository struct {
	db *gorm.DB
}

// NewClickRepository creates a new instance of ClickRepository
func NewClickRepository(db *gorm.DB) repository.ClickRepository {
	return &clickRepository{db: db}
}

func (r *clickRepository) Create(ctx context.Context, click *domain.Click) error {
	e := mapper.ClickToEntity(click)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *clickRepository) GetByLinkID(ctx context.Context, linkID string, page, pageSize int) ([]domain.Click, int64, error) {
	var entities []*entity.ClickEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ClickEntity{}).Where("link_id = ?", linkID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Order by clicked_at descending
	query = query.Order("clicked_at DESC")

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	// Convert entities to domain
	clicks := make([]domain.Click, len(entities))
	for i, e := range entities {
		clicks[i] = *mapper.ClickToDomain(e)
	}

	return clicks, total, nil
}

func (r *clickRepository) GetByDateRange(ctx context.Context, linkID string, startDate, endDate time.Time, page, pageSize int) ([]domain.Click, int64, error) {
	var entities []*entity.ClickEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ClickEntity{}).
		Where("link_id = ? AND clicked_at >= ? AND clicked_at <= ?", linkID, startDate, endDate)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Order by clicked_at descending
	query = query.Order("clicked_at DESC")

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	// Convert entities to domain
	clicks := make([]domain.Click, len(entities))
	for i, e := range entities {
		clicks[i] = *mapper.ClickToDomain(e)
	}

	return clicks, total, nil
}

func (r *clickRepository) GetStats(ctx context.Context, linkID string) (*repository.ClickStats, error) {
	stats := &repository.ClickStats{
		LinkID:          linkID,
		ClicksByCountry: make(map[string]int),
		ClicksByDevice:  make(map[string]int),
		TopReferers:     make([]string, 0),
	}

	// Get total clicks
	var totalClicks int64
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
		Where("link_id = ?", linkID).
		Count(&totalClicks).Error; err != nil {
		return nil, err
	}
	stats.TotalClicks = int(totalClicks)

	// Get unique IPs
	var uniqueIPs int64
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
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
		Model(&entity.ClickEntity{}).
		Where("link_id = ? AND clicked_at >= ?", linkID, today).
		Count(&clicksToday).Error; err != nil {
		return nil, err
	}
	stats.ClicksToday = int(clicksToday)

	// Get clicks this week
	weekStart := time.Now().AddDate(0, 0, -7)
	var clicksWeek int64
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
		Where("link_id = ? AND clicked_at >= ?", linkID, weekStart).
		Count(&clicksWeek).Error; err != nil {
		return nil, err
	}
	stats.ClicksThisWeek = int(clicksWeek)

	// Get clicks this month
	monthStart := time.Now().AddDate(0, -1, 0)
	var clicksMonth int64
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
		Where("link_id = ? AND clicked_at >= ?", linkID, monthStart).
		Count(&clicksMonth).Error; err != nil {
		return nil, err
	}
	stats.ClicksThisMonth = int(clicksMonth)

	// Get clicks by country
	type CountryCount struct {
		Country string
		Count   int64
	}
	var countryCounts []CountryCount
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
		Select("country, COUNT(*) as count").
		Where("link_id = ? AND country IS NOT NULL", linkID).
		Group("country").
		Order("count DESC").
		Limit(10).
		Scan(&countryCounts).Error; err != nil {
		return nil, err
	}
	for _, cc := range countryCounts {
		stats.ClicksByCountry[cc.Country] = int(cc.Count)
	}

	// Get clicks by device
	type DeviceCount struct {
		DeviceType string
		Count      int64
	}
	var deviceCounts []DeviceCount
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
		Select("device_type, COUNT(*) as count").
		Where("link_id = ? AND device_type IS NOT NULL", linkID).
		Group("device_type").
		Order("count DESC").
		Scan(&deviceCounts).Error; err != nil {
		return nil, err
	}
	for _, dc := range deviceCounts {
		stats.ClicksByDevice[dc.DeviceType] = int(dc.Count)
	}

	// Get top referers
	type RefererCount struct {
		Referer string
		Count   int64
	}
	var refererCounts []RefererCount
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
		Select("referer, COUNT(*) as count").
		Where("link_id = ? AND referer IS NOT NULL AND referer != ''", linkID).
		Group("referer").
		Order("count DESC").
		Limit(10).
		Scan(&refererCounts).Error; err != nil {
		return nil, err
	}
	for _, rc := range refererCounts {
		stats.TopReferers = append(stats.TopReferers, rc.Referer)
	}

	return stats, nil
}

func (r *clickRepository) GetClicksByDate(ctx context.Context, linkID string, startDate, endDate time.Time) (map[string]int, error) {
	type DateCount struct {
		Date  string
		Count int64
	}

	var dateCounts []DateCount
	if err := r.db.WithContext(ctx).
		Model(&entity.ClickEntity{}).
		Select("DATE(clicked_at) as date, COUNT(*) as count").
		Where("link_id = ? AND clicked_at >= ? AND clicked_at <= ?", linkID, startDate, endDate).
		Group("DATE(clicked_at)").
		Order("date ASC").
		Scan(&dateCounts).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, dc := range dateCounts {
		result[dc.Date] = int(dc.Count)
	}

	return result, nil
}
