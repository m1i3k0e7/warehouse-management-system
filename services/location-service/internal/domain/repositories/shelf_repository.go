package repositories

import (
	"context"

	"warehouse/location-service/internal/domain/entities"
)

type ShelfRepository interface {
	CreateShelf(ctx context.Context, shelf *entities.Shelf) error
	GetShelfByID(ctx context.Context, id string) (*entities.Shelf, error)
	GetAllShelves(ctx context.Context) ([]entities.Shelf, error)
	GetShelvesByZone(ctx context.Context, zone string) ([]entities.Shelf, error)
	UpdateShelf(ctx context.Context, shelf *entities.Shelf) error
	DeleteShelf(ctx context.Context, id string) error

	CreateSlot(ctx context.Context, slot *entities.Slot) error
	GetSlotByID(ctx context.Context, id string) (*entities.Slot, error)
	UpdateSlot(ctx context.Context, slot *entities.Slot) error
	DeleteSlot(ctx context.Context, id string) error
	GetSlotsByShelfID(ctx context.Context, shelfID string) ([]entities.Slot, error)
}
