package services

import (
	"context"
	"fmt"

	"warehouse/location-service/internal/domain/entities"
	"warehouse/location-service/internal/domain/repositories"
)

type AllocationService struct {
	shelfRepo repositories.ShelfRepository
	// Add other dependencies like cache, etc.
}

func NewAllocationService(shelfRepo repositories.ShelfRepository) *AllocationService {
	return &AllocationService{
		shelfRepo: shelfRepo,
	}
}

// AllocateSlot finds an optimal slot for a given material type and zone.
// This is a simplified example; real-world allocation would be much more complex.
func (s *AllocationService) AllocateSlot(ctx context.Context, materialType, zone string) (*entities.Slot, error) {
	// For simplicity, let's just find the first empty slot in any shelf within the zone
	// In a real system, this would involve complex algorithms (e.g., ABC analysis, FIFO, LIFO)

	shelves, err := s.shelfRepo.GetShelvesByZone(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelves by zone: %w", err)
	}

	for _, shelf := range shelves {
		for _, slot := range shelf.Slots {
			if slot.Status == "empty" { // Assuming "empty" status
				// Further checks for material type compatibility, capacity, etc. would go here
				return &slot, nil
			}
		}
	}

	return nil, fmt.Errorf("no empty slot found in zone %s for material type %s", zone, materialType)
}
