package mapper

import (
	"time"

	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository/entity"
)

func CharacterToDomain(e *entity.Character) *domain.Character {
	if e == nil {
		return nil
	}

	var deletedAt *time.Time
	if e.DeletedAt.Valid {
		deletedAt = &e.DeletedAt.Time
	}

	return &domain.Character{
		ID:        e.ID,
		Name:      e.Name,
		Realm:     e.Realm,
		Region:    e.Region,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func CharacterToEntity(d *domain.Character) *entity.Character {
	if d == nil {
		return nil
	}

	return &entity.Character{
		ID:        d.ID,
		Name:      d.Name,
		Realm:     d.Realm,
		Region:    d.Region,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func CharacterDetailsToDomain(e *entity.CharacterDetails) *domain.CharacterDetails {
	if e == nil {
		return nil
	}

	return &domain.CharacterDetails{
		ID:                e.ID,
		CharacterID:       e.CharacterID,
		DisplayName:       e.DisplayName,
		DisplayRealm:      e.DisplayRealm,
		Level:             e.Level,
		ItemLevel:         e.ItemLevel,
		ClassID:           e.ClassID,
		RaceID:            e.RaceID,
		FactionID:         e.FactionID,
		GuildID:           e.GuildID,
		ThumbnailURL:      e.ThumbnailURL,
		AchievementPoints: e.AchievementPoints,
		LastSyncedAt:      e.LastSyncedAt,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}

func CharacterDetailsToEntity(d *domain.CharacterDetails) *entity.CharacterDetails {
	if d == nil {
		return nil
	}

	return &entity.CharacterDetails{
		ID:                d.ID,
		CharacterID:       d.CharacterID,
		DisplayName:       d.DisplayName,
		DisplayRealm:      d.DisplayRealm,
		Level:             d.Level,
		ItemLevel:         d.ItemLevel,
		ClassID:           d.ClassID,
		RaceID:            d.RaceID,
		FactionID:         d.FactionID,
		GuildID:           d.GuildID,
		ThumbnailURL:      d.ThumbnailURL,
		AchievementPoints: d.AchievementPoints,
		LastSyncedAt:      d.LastSyncedAt,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

func GuildToDomain(e *entity.Guild) *domain.Guild {
	if e == nil {
		return nil
	}

	var deletedAt *time.Time
	if e.DeletedAt.Valid {
		deletedAt = &e.DeletedAt.Time
	}

	return &domain.Guild{
		ID:                e.ID,
		Name:              e.Name,
		Realm:             e.Realm,
		Region:            e.Region,
		FactionID:         e.FactionID,
		MemberCount:       e.MemberCount,
		AchievementPoints: e.AchievementPoints,
		LastSyncedAt:      e.LastSyncedAt,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
		DeletedAt:         deletedAt,
	}
}

func GuildToEntity(d *domain.Guild) *entity.Guild {
	if d == nil {
		return nil
	}

	return &entity.Guild{
		ID:                d.ID,
		Name:              d.Name,
		Realm:             d.Realm,
		Region:            d.Region,
		FactionID:         d.FactionID,
		MemberCount:       d.MemberCount,
		AchievementPoints: d.AchievementPoints,
		LastSyncedAt:      d.LastSyncedAt,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}
