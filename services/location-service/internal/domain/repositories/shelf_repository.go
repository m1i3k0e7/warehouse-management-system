package repositories

import (
	"context"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/entities"
)

// ShelfRepository defines the interface for interacting with shelf storage.
type ShelfRepository interface {
	FindByID(ctx context.Context, id string) (*entities.Shelf, error)
	Save(ctx context.Context, shelf *entities.Shelf) error
	UpdateSlotStatus(ctx context.Context, shelfID string, slotID string, status entities.SlotStatus, materialID string) error
}