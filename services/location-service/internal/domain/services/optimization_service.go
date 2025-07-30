package services

import (
	"context"
	"fmt"

	"warehouse/location-service/internal/domain/entities"
	"warehouse/location-service/internal/domain/repositories"
)

type OptimizationService struct {
	shelfRepo repositories.ShelfRepository
	// Add other dependencies like analytics service client
}

func NewOptimizationService(shelfRepo repositories.ShelfRepository) *OptimizationService {
	return &OptimizationService{
		shelfRepo: shelfRepo,
	}
}

// SuggestOptimalStorage suggests optimal storage locations based on various criteria.
// This is a simplified example.
func (s *OptimizationService) SuggestOptimalStorage(ctx context.Context, materialType string) ([]entities.Slot, error) {
	// In a real system, this would involve analyzing historical data (from Analytics Service),
	// material properties, access frequency, and current warehouse occupancy.

	// For simplicity, let's suggest empty slots on the first shelf found.
	shelves, err := s.shelfRepo.GetAllShelves(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all shelves: %w", err)
	}

	if len(shelves) == 0 {
		return nil, fmt.Errorf("no shelves found for optimization")
	}

	var optimalSlots []entities.Slot
	for _, shelf := range shelves {
		for _, slot := range shelf.Slots {
			if slot.Status == "empty" { // Assuming "empty" status
				optimalSlots = append(optimalSlots, slot)
				if len(optimalSlots) >= 5 { // Suggest up to 5 optimal slots
					return optimalSlots, nil
				}
			}
		}
	}

	return optimalSlots, nil
}
