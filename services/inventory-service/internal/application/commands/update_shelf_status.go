package commands

import (
	"context"
	"WMS/services/inventory-service/internal/domain/services"
)

type UpdateShelfStatusCommand struct {
	ShelfID string
	Status  string
}

type UpdateShelfStatusCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewUpdateShelfStatusCommandHandler(inventoryService *services.InventoryService) *UpdateShelfStatusCommandHandler {
	return &UpdateShelfStatusCommandHandler{inventoryService: inventoryService}
}

func (h *UpdateShelfStatusCommandHandler) Handle(ctx context.Context, cmd UpdateShelfStatusCommand) error {
	return h.inventoryService.UpdateShelfStatus(ctx, cmd.ShelfID, cmd.Status)
}
