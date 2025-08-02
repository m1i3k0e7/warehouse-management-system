package services

import (
	"context"

	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/entities"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/repositories"
)

// AllocationService provides logic for allocating slots for materials.
	ype AllocationService struct {
	layoutRepo repositories.LayoutRepository
}

// NewAllocationService creates a new AllocationService.
func NewAllocationService(layoutRepo repositories.LayoutRepository) *AllocationService {
	return &AllocationService{layoutRepo: layoutRepo}
}

// SuggestSlot finds and suggests a suitable slot for a given material type.
func (s *AllocationService) SuggestSlot(ctx context.Context, materialType, zoneID string) (*entities.Shelf, *entities.Slot, error) {
	// This is a simple first-fit algorithm. More complex logic can be added later.
	// e.g., based on material type, historical data, or proximity to other materials.

	shelves, err := s.layoutRepo.FindAllShelvesInZone(ctx, zoneID)
	if err != nil {
		return nil, nil, err
	}

	for _, shelf := range shelves {
		for _, slot := range shelf.Slots {
			if slot.Status == entities.StatusEmpty {
				return shelf, &slot, nil
			}
		}
	}

	return nil, nil, nil // No available slot found
}