package commands

import (
	"context"
	"WMS/services/inventory-service/internal/domain/services"
)

type BatchPlaceMaterialsCommand struct {
	Commands []PlaceMaterialCommand
}

type BatchPlaceMaterialsCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewBatchPlaceMaterialsCommandHandler(inventoryService *services.InventoryService) *BatchPlaceMaterialsCommandHandler {
	return &BatchPlaceMaterialsCommandHandler{inventoryService: inventoryService}
}

func (h *BatchPlaceMaterialsCommandHandler) Handle(ctx context.Context, cmd BatchPlaceMaterialsCommand) error {
	params := make([]services.PlaceMaterialParams, len(cmd.Commands))
	for i, c := range cmd.Commands {
		params[i] = services.PlaceMaterialParams{
			MaterialBarcode: c.MaterialBarcode,
			SlotID:          c.SlotID,
			OperatorID:      c.OperatorID,
		}
	}
	return h.inventoryService.BatchPlaceMaterials(ctx, params)
}
