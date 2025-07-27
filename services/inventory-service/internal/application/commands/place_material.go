package commands

import (
	"context"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/repositories"
	"warehouse/internal/domain/services"
	"warehouse/pkg/errors"
)

type PlaceMaterialCommand struct {
	MaterialBarcode string
	SlotID          string
	OperatorID      string
}

type PlaceMaterialCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewPlaceMaterialCommandHandler(inventoryService *services.InventoryService) *PlaceMaterialCommandHandler {
	return &PlaceMaterialCommandHandler{inventoryService: inventoryService}
}

func (h *PlaceMaterialCommandHandler) Handle(ctx context.Context, cmd PlaceMaterialCommand) error {
	return h.inventoryService.PlaceMaterial(ctx, services.PlaceMaterialCommand{
		MaterialBarcode: cmd.MaterialBarcode,
		SlotID:          cmd.SlotID,
		OperatorID:      cmd.OperatorID,
	})
}
