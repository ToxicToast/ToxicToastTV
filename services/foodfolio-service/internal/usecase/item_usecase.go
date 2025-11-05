package usecase

import (
	"context"
	"errors"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

var (
	ErrItemNotFound    = errors.New("item not found")
	ErrInvalidItemData = errors.New("invalid item data")
)

type ItemUseCase interface {
	CreateItem(ctx context.Context, name, categoryID, companyID, typeID string) (*domain.Item, error)
	GetItemByID(ctx context.Context, id string) (*domain.Item, error)
	GetItemBySlug(ctx context.Context, slug string) (*domain.Item, error)
	GetItemWithVariants(ctx context.Context, id string, includeDetails bool) (*domain.Item, error)
	ListItems(ctx context.Context, page, pageSize int, categoryID, companyID, typeID, search *string, includeDeleted bool) ([]*domain.Item, int64, error)
	SearchItems(ctx context.Context, query string, page, pageSize int, categoryID, companyID *string) ([]*domain.Item, int64, error)
	UpdateItem(ctx context.Context, id, name, categoryID, companyID, typeID string) (*domain.Item, error)
	DeleteItem(ctx context.Context, id string) error
}

type itemUseCase struct {
	itemRepo     interfaces.ItemRepository
	categoryRepo interfaces.CategoryRepository
	companyRepo  interfaces.CompanyRepository
	typeRepo     interfaces.TypeRepository
}

func NewItemUseCase(
	itemRepo interfaces.ItemRepository,
	categoryRepo interfaces.CategoryRepository,
	companyRepo interfaces.CompanyRepository,
	typeRepo interfaces.TypeRepository,
) ItemUseCase {
	return &itemUseCase{
		itemRepo:     itemRepo,
		categoryRepo: categoryRepo,
		companyRepo:  companyRepo,
		typeRepo:     typeRepo,
	}
}

func (uc *itemUseCase) CreateItem(ctx context.Context, name, categoryID, companyID, typeID string) (*domain.Item, error) {
	// Validate
	if name == "" || categoryID == "" || companyID == "" || typeID == "" {
		return nil, ErrInvalidItemData
	}

	// Validate category exists
	category, err := uc.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	// Validate company exists
	company, err := uc.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, errors.New("company not found")
	}

	// Validate type exists
	typeEntity, err := uc.typeRepo.GetByID(ctx, typeID)
	if err != nil {
		return nil, err
	}
	if typeEntity == nil {
		return nil, errors.New("type not found")
	}

	item := &domain.Item{
		Name:       name,
		CategoryID: categoryID,
		CompanyID:  companyID,
		TypeID:     typeID,
	}

	if err := uc.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (uc *itemUseCase) GetItemByID(ctx context.Context, id string) (*domain.Item, error) {
	item, err := uc.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, ErrItemNotFound
	}

	return item, nil
}

func (uc *itemUseCase) GetItemBySlug(ctx context.Context, slug string) (*domain.Item, error) {
	item, err := uc.itemRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, ErrItemNotFound
	}

	return item, nil
}

func (uc *itemUseCase) GetItemWithVariants(ctx context.Context, id string, includeDetails bool) (*domain.Item, error) {
	item, err := uc.itemRepo.GetWithVariants(ctx, id, includeDetails)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, ErrItemNotFound
	}

	return item, nil
}

func (uc *itemUseCase) ListItems(ctx context.Context, page, pageSize int, categoryID, companyID, typeID, search *string, includeDeleted bool) ([]*domain.Item, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.itemRepo.List(ctx, offset, pageSize, categoryID, companyID, typeID, search, includeDeleted)
}

func (uc *itemUseCase) SearchItems(ctx context.Context, query string, page, pageSize int, categoryID, companyID *string) ([]*domain.Item, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	if query == "" {
		return nil, 0, errors.New("search query cannot be empty")
	}

	return uc.itemRepo.Search(ctx, query, offset, pageSize, categoryID, companyID)
}

func (uc *itemUseCase) UpdateItem(ctx context.Context, id, name, categoryID, companyID, typeID string) (*domain.Item, error) {
	item, err := uc.GetItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate
	if name == "" || categoryID == "" || companyID == "" || typeID == "" {
		return nil, ErrInvalidItemData
	}

	// Validate category exists
	category, err := uc.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	// Validate company exists
	company, err := uc.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, errors.New("company not found")
	}

	// Validate type exists
	typeEntity, err := uc.typeRepo.GetByID(ctx, typeID)
	if err != nil {
		return nil, err
	}
	if typeEntity == nil {
		return nil, errors.New("type not found")
	}

	item.Name = name
	item.CategoryID = categoryID
	item.CompanyID = companyID
	item.TypeID = typeID

	if err := uc.itemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (uc *itemUseCase) DeleteItem(ctx context.Context, id string) error {
	_, err := uc.GetItemByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.itemRepo.Delete(ctx, id)
}
