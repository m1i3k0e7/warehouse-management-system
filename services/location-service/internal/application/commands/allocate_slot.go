package commands

import (
	"context"
	"fmt"

	"warehouse/location-service/internal/domain/entities"
	"warehouse/location-service/internal/domain/services"
)

type AllocateSlotCommand struct {
	MaterialType string `json:"material_type"`
	Zone         string `json:"zone"`
}

type AllocateSlotCommandHandler struct {
	allocationService *services.AllocationService
}

func NewAllocateSlotCommandHandler(allocationService *services.AllocationService) *AllocateSlotCommandHandler {
	return &AllocateSlotCommandHandler{
		allocationService: allocationService,
	}
}

func (h *AllocateSlotCommandHandler) Handle(ctx context.Context, cmd AllocateSlotCommand) (*entities.Slot, error) {
	slot, err := h.allocationService.AllocateSlot(ctx, cmd.MaterialType, cmd.Zone)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate slot: %w", err)
	}
	return slot, nil
}
