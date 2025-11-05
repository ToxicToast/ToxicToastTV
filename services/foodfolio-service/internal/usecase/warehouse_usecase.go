package usecase

import (
	"context"
	"errors"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

var (
	ErrWarehouseNotFound    = errors.New("warehouse not found")
	ErrInvalidWarehouseData = errors.New("invalid warehouse data")
)

type WarehouseUseCase interface {
	CreateWarehouse(ctx context.Context, name string) (*domain.Warehouse, error)
	GetWarehouseByID(ctx context.Context, id string) (*domain.Warehouse, error)
	GetWarehouseBySlug(ctx context.Context, slug string) (*domain.Warehouse, error)
	ListWarehouses(ctx context.Context, page, pageSize int, search string, includeDeleted bool) ([]*domain.Warehouse, int64, error)
	UpdateWarehouse(ctx context.Context, id, name string) (*domain.Warehouse, error)
	DeleteWarehouse(ctx context.Context, id string) error
}

type warehouseUseCase struct {
	warehouseRepo interfaces.WarehouseRepository
}

func NewWarehouseUseCase(warehouseRepo interfaces.WarehouseRepository) WarehouseUseCase {
	return &warehouseUseCase{
		warehouseRepo: warehouseRepo,
	}
}

func (uc *warehouseUseCase) CreateWarehouse(ctx context.Context, name string) (*domain.Warehouse, error) {
	if name == "" {
		return nil, ErrInvalidWarehouseData
	}

	warehouse := &domain.Warehouse{
		Name: name,
	}

	if err := uc.warehouseRepo.Create(ctx, warehouse); err != nil {
		return nil, err
	}

	return warehouse, nil
}

func (uc *warehouseUseCase) GetWarehouseByID(ctx context.Context, id string) (*domain.Warehouse, error) {
	warehouse, err := uc.warehouseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if warehouse == nil {
		return nil, ErrWarehouseNotFound
	}

	return warehouse, nil
}

func (uc *warehouseUseCase) GetWarehouseBySlug(ctx context.Context, slug string) (*domain.Warehouse, error) {
	warehouse, err := uc.warehouseRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if warehouse == nil {
		return nil, ErrWarehouseNotFound
	}

	return warehouse, nil
}

func (uc *warehouseUseCase) ListWarehouses(ctx context.Context, page, pageSize int, search string, includeDeleted bool) ([]*domain.Warehouse, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.warehouseRepo.List(ctx, offset, pageSize, search, includeDeleted)
}

func (uc *warehouseUseCase) UpdateWarehouse(ctx context.Context, id, name string) (*domain.Warehouse, error) {
	warehouse, err := uc.GetWarehouseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, ErrInvalidWarehouseData
	}

	warehouse.Name = name

	if err := uc.warehouseRepo.Update(ctx, warehouse); err != nil {
		return nil, err
	}

	return warehouse, nil
}

func (uc *warehouseUseCase) DeleteWarehouse(ctx context.Context, id string) error {
	_, err := uc.GetWarehouseByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.warehouseRepo.Delete(ctx, id)
}
