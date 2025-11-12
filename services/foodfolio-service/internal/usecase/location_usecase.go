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
	ErrLocationNotFound    = errors.New("location not found")
	ErrInvalidLocationData = errors.New("invalid location data")
)

type LocationUseCase interface {
	CreateLocation(ctx context.Context, name string, parentID *string) (*domain.Location, error)
	GetLocationByID(ctx context.Context, id string) (*domain.Location, error)
	GetLocationBySlug(ctx context.Context, slug string) (*domain.Location, error)
	ListLocations(ctx context.Context, page, pageSize int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Location, int64, error)
	GetLocationTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Location, error)
	GetRootLocations(ctx context.Context) ([]*domain.Location, error)
	UpdateLocation(ctx context.Context, id, name string, parentID *string) (*domain.Location, error)
	DeleteLocation(ctx context.Context, id string) error
}

type locationUseCase struct {
	locationRepo  interfaces.LocationRepository
	kafkaProducer *kafka.Producer
}

func NewLocationUseCase(locationRepo interfaces.LocationRepository, kafkaProducer *kafka.Producer) LocationUseCase {
	return &locationUseCase{
		locationRepo:  locationRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (uc *locationUseCase) CreateLocation(ctx context.Context, name string, parentID *string) (*domain.Location, error) {
	if name == "" {
		return nil, ErrInvalidLocationData
	}

	// Validate parent exists if provided
	if parentID != nil && *parentID != "" {
		parent, err := uc.locationRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New("parent location not found")
		}
	}

	location := &domain.Location{
		Name:     name,
		ParentID: parentID,
	}

	if err := uc.locationRepo.Create(ctx, location); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioLocationCreatedEvent{
			LocationID: location.ID,
			Name:       location.Name,
			Slug:       location.Slug,
			ParentID:   location.ParentID,
			CreatedAt:  time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioLocationCreated("foodfolio.location.created", event); err != nil {
			log.Printf("Warning: Failed to publish location created event: %v", err)
		}
	}

	return location, nil
}

func (uc *locationUseCase) GetLocationByID(ctx context.Context, id string) (*domain.Location, error) {
	location, err := uc.locationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if location == nil {
		return nil, ErrLocationNotFound
	}

	return location, nil
}

func (uc *locationUseCase) GetLocationBySlug(ctx context.Context, slug string) (*domain.Location, error) {
	location, err := uc.locationRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if location == nil {
		return nil, ErrLocationNotFound
	}

	return location, nil
}

func (uc *locationUseCase) ListLocations(ctx context.Context, page, pageSize int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Location, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.locationRepo.List(ctx, offset, pageSize, parentID, includeChildren, includeDeleted)
}

func (uc *locationUseCase) GetLocationTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Location, error) {
	// Validate root exists if provided
	if rootID != nil && *rootID != "" {
		root, err := uc.locationRepo.GetByID(ctx, *rootID)
		if err != nil {
			return nil, err
		}
		if root == nil {
			return nil, errors.New("root location not found")
		}
	}

	return uc.locationRepo.GetTree(ctx, rootID, maxDepth)
}

func (uc *locationUseCase) GetRootLocations(ctx context.Context) ([]*domain.Location, error) {
	return uc.locationRepo.GetRootLocations(ctx)
}

func (uc *locationUseCase) UpdateLocation(ctx context.Context, id, name string, parentID *string) (*domain.Location, error) {
	location, err := uc.GetLocationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, ErrInvalidLocationData
	}

	// Check for circular reference
	if parentID != nil && *parentID != "" {
		if *parentID == id {
			return nil, ErrCircularReference
		}

		// Validate parent exists
		parent, err := uc.locationRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New("parent location not found")
		}

		// Check if new parent is a child of this location
		if parent.ParentID != nil && *parent.ParentID == id {
			return nil, ErrCircularReference
		}
	}

	location.Name = name
	location.ParentID = parentID

	if err := uc.locationRepo.Update(ctx, location); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioLocationUpdatedEvent{
			LocationID: location.ID,
			Name:       location.Name,
			Slug:       location.Slug,
			ParentID:   location.ParentID,
			UpdatedAt:  time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioLocationUpdated("foodfolio.location.updated", event); err != nil {
			log.Printf("Warning: Failed to publish location updated event: %v", err)
		}
	}

	return location, nil
}

func (uc *locationUseCase) DeleteLocation(ctx context.Context, id string) error {
	_, err := uc.GetLocationByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if location has children
	children, err := uc.locationRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}

	if len(children) > 0 {
		return errors.New("cannot delete location with children")
	}

	if err := uc.locationRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioLocationDeletedEvent{
			LocationID: id,
			DeletedAt:  time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioLocationDeleted("foodfolio.location.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish location deleted event: %v", err)
		}
	}

	return nil
}
