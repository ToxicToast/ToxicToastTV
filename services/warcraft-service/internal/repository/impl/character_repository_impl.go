package impl

import (
	"context"

	"gorm.io/gorm"
	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository"
	"toxictoast/services/warcraft-service/internal/repository/entity"
	"toxictoast/services/warcraft-service/internal/repository/mapper"
)

type characterRepositoryImpl struct {
	db *gorm.DB
}

func NewCharacterRepository(db *gorm.DB) repository.CharacterRepository {
	return &characterRepositoryImpl{db: db}
}

func (r *characterRepositoryImpl) Create(ctx context.Context, character *domain.Character) (*domain.Character, error) {
	e := mapper.CharacterToEntity(character)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return mapper.CharacterToDomain(e), nil
}

func (r *characterRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Character, error) {
	var e entity.Character
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.CharacterToDomain(&e), nil
}

func (r *characterRepositoryImpl) FindByNameRealmRegion(ctx context.Context, name, realm, region string) (*domain.Character, error) {
	var e entity.Character
	if err := r.db.WithContext(ctx).Where("name = ? AND realm = ? AND region = ?", name, realm, region).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.CharacterToDomain(&e), nil
}

func (r *characterRepositoryImpl) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.Character, int, error) {
	var entities []entity.Character
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Character{})

	// Apply filters
	for key, value := range filters {
		if value != nil {
			query = query.Where(key+" = ?", value)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	characters := make([]*domain.Character, len(entities))
	for i, e := range entities {
		characters[i] = mapper.CharacterToDomain(&e)
	}

	return characters, int(total), nil
}

func (r *characterRepositoryImpl) Update(ctx context.Context, character *domain.Character) (*domain.Character, error) {
	e := mapper.CharacterToEntity(character)
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return nil, err
	}
	return mapper.CharacterToDomain(e), nil
}

func (r *characterRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Character{}, "id = ?", id).Error
}

// Guild Repository

type guildRepositoryImpl struct {
	db *gorm.DB
}

func NewGuildRepository(db *gorm.DB) repository.GuildRepository {
	return &guildRepositoryImpl{db: db}
}

func (r *guildRepositoryImpl) Create(ctx context.Context, guild *domain.Guild) (*domain.Guild, error) {
	e := mapper.GuildToEntity(guild)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return mapper.GuildToDomain(e), nil
}

func (r *guildRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Guild, error) {
	var e entity.Guild
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.GuildToDomain(&e), nil
}

func (r *guildRepositoryImpl) FindByNameRealmRegion(ctx context.Context, name, realm, region string) (*domain.Guild, error) {
	var e entity.Guild
	if err := r.db.WithContext(ctx).Where("name = ? AND realm = ? AND region = ?", name, realm, region).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.GuildToDomain(&e), nil
}

func (r *guildRepositoryImpl) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.Guild, int, error) {
	var entities []entity.Guild
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Guild{})

	// Apply filters
	for key, value := range filters {
		if value != nil {
			query = query.Where(key+" = ?", value)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	guilds := make([]*domain.Guild, len(entities))
	for i, e := range entities {
		guilds[i] = mapper.GuildToDomain(&e)
	}

	return guilds, int(total), nil
}

func (r *guildRepositoryImpl) Update(ctx context.Context, guild *domain.Guild) (*domain.Guild, error) {
	e := mapper.GuildToEntity(guild)
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return nil, err
	}
	return mapper.GuildToDomain(e), nil
}

func (r *guildRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Guild{}, "id = ?", id).Error
}

// CharacterDetails Repository

type characterDetailsRepositoryImpl struct {
	db *gorm.DB
}

func NewCharacterDetailsRepository(db *gorm.DB) repository.CharacterDetailsRepository {
	return &characterDetailsRepositoryImpl{db: db}
}

func (r *characterDetailsRepositoryImpl) Create(ctx context.Context, details *domain.CharacterDetails) (*domain.CharacterDetails, error) {
	e := mapper.CharacterDetailsToEntity(details)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return mapper.CharacterDetailsToDomain(e), nil
}

func (r *characterDetailsRepositoryImpl) FindByCharacterID(ctx context.Context, characterID string) (*domain.CharacterDetails, error) {
	var e entity.CharacterDetails
	if err := r.db.WithContext(ctx).Where("character_id = ?", characterID).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.CharacterDetailsToDomain(&e), nil
}

func (r *characterDetailsRepositoryImpl) Update(ctx context.Context, details *domain.CharacterDetails) (*domain.CharacterDetails, error) {
	e := mapper.CharacterDetailsToEntity(details)
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return nil, err
	}
	return mapper.CharacterDetailsToDomain(e), nil
}

func (r *characterDetailsRepositoryImpl) Delete(ctx context.Context, characterID string) error {
	return r.db.WithContext(ctx).Delete(&entity.CharacterDetails{}, "character_id = ?", characterID).Error
}
