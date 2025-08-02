package commands

import (
	"context"

	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/entities"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/repositories"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/services"
)

// AllocateSlotCommandHandler handles the AllocateSlot command.
type AllocateSlotCommandHandler struct {
	allocationService *services.AllocationService
	shelfRepo         repositories.ShelfRepository
}

// NewAllocateSlotCommandHandler creates a new AllocateSlotCommandHandler.
func NewAllocateSlotCommandHandler(allocationService *services.AllocationService, shelfRepo repositories.ShelfRepository) *AllocateSlotCommandHandler {
	return &AllocateSlotCommandHandler{
		allocationService: allocationService,
		shelfRepo:         shelfRepo,
	}
}

// Handle executes the command.
func (h *AllocateSlotCommandHandler) Handle(ctx context.Context, materialType, zoneID, materialID string) (*entities.Shelf, *entities.Slot, error) {
	shelf, slot, err := h.allocationService.SuggestSlot(ctx, materialType, zoneID)
	if err != nil {
		return nil, nil, err
	}
	if shelf == nil || slot == nil {
		return nil, nil, nil // No slot found
	}

	err = h.shelfRepo.UpdateSlotStatus(ctx, shelf.ID, slot.ID, entities.StatusReserved, materialID)
	if err != nil {
		return nil, nil, err
	}

	slot.Status = entities.StatusReserved
	slot.MaterialID = materialID

	return shelf, slot, nil
}