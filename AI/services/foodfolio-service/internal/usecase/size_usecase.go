package usecase

import (
	"context"
	"errors"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

var (
	ErrSizeNotFound    = errors.New("size not found")
	ErrInvalidSizeData = errors.New("invalid size data")
)

type SizeUseCase interface {
	CreateSize(ctx context.Context, name string, value float64, unit string) (*domain.Size, error)
	GetSizeByID(ctx context.Context, id string) (*domain.Size, error)
	GetSizeByName(ctx context.Context, name string) (*domain.Size, error)
	ListSizes(ctx context.Context, page, pageSize int, unit string, minValue, maxValue *float64, includeDeleted bool) ([]*domain.Size, int64, error)
	UpdateSize(ctx context.Context, id, name string, value float64, unit string) (*domain.Size, error)
	DeleteSize(ctx context.Context, id string) error
}

type sizeUseCase struct {
	sizeRepo interfaces.SizeRepository
}

func NewSizeUseCase(sizeRepo interfaces.SizeRepository) SizeUseCase {
	return &sizeUseCase{
		sizeRepo: sizeRepo,
	}
}

func (uc *sizeUseCase) CreateSize(ctx context.Context, name string, value float64, unit string) (*domain.Size, error) {
	// Validate
	if name == "" || unit == "" || value <= 0 {
		return nil, ErrInvalidSizeData
	}

	size := &domain.Size{
		Name:  name,
		Value: value,
		Unit:  unit,
	}

	if err := uc.sizeRepo.Create(ctx, size); err != nil {
		return nil, err
	}

	return size, nil
}

func (uc *sizeUseCase) GetSizeByID(ctx context.Context, id string) (*domain.Size, error) {
	size, err := uc.sizeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if size == nil {
		return nil, ErrSizeNotFound
	}

	return size, nil
}

func (uc *sizeUseCase) GetSizeByName(ctx context.Context, name string) (*domain.Size, error) {
	size, err := uc.sizeRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if size == nil {
		return nil, ErrSizeNotFound
	}

	return size, nil
}

func (uc *sizeUseCase) ListSizes(ctx context.Context, page, pageSize int, unit string, minValue, maxValue *float64, includeDeleted bool) ([]*domain.Size, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.sizeRepo.List(ctx, offset, pageSize, unit, minValue, maxValue, includeDeleted)
}

func (uc *sizeUseCase) UpdateSize(ctx context.Context, id, name string, value float64, unit string) (*domain.Size, error) {
	size, err := uc.GetSizeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate
	if name == "" || unit == "" || value <= 0 {
		return nil, ErrInvalidSizeData
	}

	size.Name = name
	size.Value = value
	size.Unit = unit

	if err := uc.sizeRepo.Update(ctx, size); err != nil {
		return nil, err
	}

	return size, nil
}

func (uc *sizeUseCase) DeleteSize(ctx context.Context, id string) error {
	_, err := uc.GetSizeByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.sizeRepo.Delete(ctx, id)
}
