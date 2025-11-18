package usecase

import (
	"context"
	"errors"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

var (
	ErrTypeNotFound      = errors.New("type not found")
	ErrTypeAlreadyExists = errors.New("type already exists")
	ErrInvalidTypeData   = errors.New("invalid type data")
)

type TypeUseCase interface {
	CreateType(ctx context.Context, name string) (*domain.Type, error)
	GetTypeByID(ctx context.Context, id string) (*domain.Type, error)
	GetTypeBySlug(ctx context.Context, slug string) (*domain.Type, error)
	ListTypes(ctx context.Context, page, pageSize int, search string, includeDeleted bool) ([]*domain.Type, int64, error)
	UpdateType(ctx context.Context, id, name string) (*domain.Type, error)
	DeleteType(ctx context.Context, id string) error
}

type typeUseCase struct {
	typeRepo interfaces.TypeRepository
}

func NewTypeUseCase(typeRepo interfaces.TypeRepository) TypeUseCase {
	return &typeUseCase{
		typeRepo: typeRepo,
	}
}

func (uc *typeUseCase) CreateType(ctx context.Context, name string) (*domain.Type, error) {
	if name == "" {
		return nil, ErrInvalidTypeData
	}

	typeEntity := &domain.Type{
		Name: name,
	}

	if err := uc.typeRepo.Create(ctx, typeEntity); err != nil {
		return nil, err
	}

	return typeEntity, nil
}

func (uc *typeUseCase) GetTypeByID(ctx context.Context, id string) (*domain.Type, error) {
	typeEntity, err := uc.typeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if typeEntity == nil {
		return nil, ErrTypeNotFound
	}

	return typeEntity, nil
}

func (uc *typeUseCase) GetTypeBySlug(ctx context.Context, slug string) (*domain.Type, error) {
	typeEntity, err := uc.typeRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if typeEntity == nil {
		return nil, ErrTypeNotFound
	}

	return typeEntity, nil
}

func (uc *typeUseCase) ListTypes(ctx context.Context, page, pageSize int, search string, includeDeleted bool) ([]*domain.Type, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.typeRepo.List(ctx, offset, pageSize, search, includeDeleted)
}

func (uc *typeUseCase) UpdateType(ctx context.Context, id, name string) (*domain.Type, error) {
	typeEntity, err := uc.GetTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, ErrInvalidTypeData
	}

	typeEntity.Name = name

	if err := uc.typeRepo.Update(ctx, typeEntity); err != nil {
		return nil, err
	}

	return typeEntity, nil
}

func (uc *typeUseCase) DeleteType(ctx context.Context, id string) error {
	_, err := uc.GetTypeByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.typeRepo.Delete(ctx, id)
}
