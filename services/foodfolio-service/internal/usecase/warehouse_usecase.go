package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/kafka"
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
	kafkaProducer *kafka.Producer
}

func NewWarehouseUseCase(warehouseRepo interfaces.WarehouseRepository, kafkaProducer *kafka.Producer) WarehouseUseCase {
	return &warehouseUseCase{
		warehouseRepo: warehouseRepo,
		kafkaProducer: kafkaProducer,
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

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioWarehouseCreatedEvent{
			WarehouseID: warehouse.ID,
			Name:        warehouse.Name,
			Slug:        warehouse.Slug,
			CreatedAt:   time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioWarehouseCreated("foodfolio.warehouse.created", event); err != nil {
			log.Printf("Warning: Failed to publish warehouse created event: %v", err)
		}
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

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioWarehouseUpdatedEvent{
			WarehouseID: warehouse.ID,
			Name:        warehouse.Name,
			Slug:        warehouse.Slug,
			UpdatedAt:   time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioWarehouseUpdated("foodfolio.warehouse.updated", event); err != nil {
			log.Printf("Warning: Failed to publish warehouse updated event: %v", err)
		}
	}

	return warehouse, nil
}

func (uc *warehouseUseCase) DeleteWarehouse(ctx context.Context, id string) error {
	_, err := uc.GetWarehouseByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.warehouseRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioWarehouseDeletedEvent{
			WarehouseID: id,
			DeletedAt:   time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioWarehouseDeleted("foodfolio.warehouse.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish warehouse deleted event: %v", err)
		}
	}

	return nil
}
