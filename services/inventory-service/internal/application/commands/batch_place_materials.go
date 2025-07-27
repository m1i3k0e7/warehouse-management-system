package commands

import (
	"context"
	"warehouse/internal/domain/services"
)

type BatchPlaceMaterialsCommand struct {
	Commands []services.PlaceMaterialCommand
}

type BatchPlaceMaterialsCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewBatchPlaceMaterialsCommandHandler(inventoryService *services.InventoryService) *BatchPlaceMaterialsCommandHandler {
	return &BatchPlaceMaterialsCommandHandler{inventoryService: inventoryService}
}

func (h *BatchPlaceMaterialsCommandHandler) Handle(ctx context.Context, cmd BatchPlaceMaterialsCommand) error {
	return h.inventoryService.BatchPlaceMaterials(ctx, cmd.Commands)
}
