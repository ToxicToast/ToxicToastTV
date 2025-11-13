package mapper

import (
	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository/entity"
)

func RaceToDomain(e *entity.Race) *domain.Race {
	if e == nil {
		return nil
	}

	return &domain.Race{
		ID:        e.ID,
		Key:       e.Key,
		Name:      e.Name,
		FactionID: e.FactionID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func RaceToEntity(d *domain.Race) *entity.Race {
	if d == nil {
		return nil
	}

	return &entity.Race{
		ID:        d.ID,
		Key:       d.Key,
		Name:      d.Name,
		FactionID: d.FactionID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func ClassToDomain(e *entity.Class) *domain.Class {
	if e == nil {
		return nil
	}

	return &domain.Class{
		ID:        e.ID,
		Key:       e.Key,
		Name:      e.Name,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func ClassToEntity(d *domain.Class) *entity.Class {
	if d == nil {
		return nil
	}

	return &entity.Class{
		ID:        d.ID,
		Key:       d.Key,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func FactionToDomain(e *entity.Faction) *domain.Faction {
	if e == nil {
		return nil
	}

	return &domain.Faction{
		ID:        e.ID,
		Key:       e.Key,
		Name:      e.Name,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func FactionToEntity(d *domain.Faction) *entity.Faction {
	if d == nil {
		return nil
	}

	return &entity.Faction{
		ID:        d.ID,
		Key:       d.Key,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
