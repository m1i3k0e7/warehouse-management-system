package commands

import (
	"context"
	"WMS/services/inventory-service/internal/domain/services"
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
	return h.inventoryService.PlaceMaterial(ctx, services.PlaceMaterialParams{
		MaterialBarcode: cmd.MaterialBarcode,
		SlotID:          cmd.SlotID,
		OperatorID:      cmd.OperatorID,
	})
}
