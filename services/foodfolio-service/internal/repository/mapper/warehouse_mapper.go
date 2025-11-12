package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// WarehouseToEntity converts domain model to database entity
func WarehouseToEntity(warehouse *domain.Warehouse) *entity.WarehouseEntity {
	if warehouse == nil {
		return nil
	}

	e := &entity.WarehouseEntity{
		ID:        warehouse.ID,
		Name:      warehouse.Name,
		Slug:      warehouse.Slug,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}

	// Convert DeletedAt
	if warehouse.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *warehouse.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// WarehouseToDomain converts database entity to domain model
func WarehouseToDomain(e *entity.WarehouseEntity) *domain.Warehouse {
	if e == nil {
		return nil
	}

	warehouse := &domain.Warehouse{
		ID:        e.ID,
		Name:      e.Name,
		Slug:      e.Slug,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		warehouse.DeletedAt = &deletedAt
	}

	return warehouse
}

// WarehousesToDomain converts slice of entities to domain models
func WarehousesToDomain(entities []*entity.WarehouseEntity) []*domain.Warehouse {
	warehouses := make([]*domain.Warehouse, 0, len(entities))
	for _, e := range entities {
		warehouses = append(warehouses, WarehouseToDomain(e))
	}
	return warehouses
}
