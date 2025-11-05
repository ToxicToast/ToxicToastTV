package usecase

import (
	"context"
	"errors"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrInvalidCategoryData = errors.New("invalid category data")
	ErrCircularReference   = errors.New("circular reference detected")
)

type CategoryUseCase interface {
	CreateCategory(ctx context.Context, name string, parentID *string) (*domain.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*domain.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error)
	ListCategories(ctx context.Context, page, pageSize int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Category, int64, error)
	GetCategoryTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Category, error)
	GetRootCategories(ctx context.Context) ([]*domain.Category, error)
	UpdateCategory(ctx context.Context, id, name string, parentID *string) (*domain.Category, error)
	DeleteCategory(ctx context.Context, id string) error
}

type categoryUseCase struct {
	categoryRepo interfaces.CategoryRepository
}

func NewCategoryUseCase(categoryRepo interfaces.CategoryRepository) CategoryUseCase {
	return &categoryUseCase{
		categoryRepo: categoryRepo,
	}
}

func (uc *categoryUseCase) CreateCategory(ctx context.Context, name string, parentID *string) (*domain.Category, error) {
	if name == "" {
		return nil, ErrInvalidCategoryData
	}

	// Validate parent exists if provided
	if parentID != nil && *parentID != "" {
		parent, err := uc.categoryRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New("parent category not found")
		}
	}

	category := &domain.Category{
		Name:     name,
		ParentID: parentID,
	}

	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (uc *categoryUseCase) GetCategoryByID(ctx context.Context, id string) (*domain.Category, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, ErrCategoryNotFound
	}

	return category, nil
}

func (uc *categoryUseCase) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	category, err := uc.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, ErrCategoryNotFound
	}

	return category, nil
}

func (uc *categoryUseCase) ListCategories(ctx context.Context, page, pageSize int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Category, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.categoryRepo.List(ctx, offset, pageSize, parentID, includeChildren, includeDeleted)
}

func (uc *categoryUseCase) GetCategoryTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Category, error) {
	// Validate root exists if provided
	if rootID != nil && *rootID != "" {
		root, err := uc.categoryRepo.GetByID(ctx, *rootID)
		if err != nil {
			return nil, err
		}
		if root == nil {
			return nil, errors.New("root category not found")
		}
	}

	return uc.categoryRepo.GetTree(ctx, rootID, maxDepth)
}

func (uc *categoryUseCase) GetRootCategories(ctx context.Context) ([]*domain.Category, error) {
	return uc.categoryRepo.GetRootCategories(ctx)
}

func (uc *categoryUseCase) UpdateCategory(ctx context.Context, id, name string, parentID *string) (*domain.Category, error) {
	category, err := uc.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, ErrInvalidCategoryData
	}

	// Check for circular reference
	if parentID != nil && *parentID != "" {
		if *parentID == id {
			return nil, ErrCircularReference
		}

		// Validate parent exists
		parent, err := uc.categoryRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New("parent category not found")
		}

		// Check if new parent is a child of this category (would create circular reference)
		if parent.ParentID != nil && *parent.ParentID == id {
			return nil, ErrCircularReference
		}
	}

	category.Name = name
	category.ParentID = parentID

	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (uc *categoryUseCase) DeleteCategory(ctx context.Context, id string) error {
	_, err := uc.GetCategoryByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if category has children
	children, err := uc.categoryRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}

	if len(children) > 0 {
		return errors.New("cannot delete category with children")
	}

	return uc.categoryRepo.Delete(ctx, id)
}
