package commands

import (
	"context"
	"WMS/services/inventory-service/internal/domain/services"
)

type RemoveMaterialCommand struct {
	SlotID     string
	OperatorID string
	Reason     string
}

type RemoveMaterialCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewRemoveMaterialCommandHandler(inventoryService *services.InventoryService) *RemoveMaterialCommandHandler {
	return &RemoveMaterialCommandHandler{inventoryService: inventoryService}
}

func (h *RemoveMaterialCommandHandler) Handle(ctx context.Context, cmd RemoveMaterialCommand) error {
	return h.inventoryService.RemoveMaterial(ctx, services.RemoveMaterialParams{
		SlotID:     cmd.SlotID,
		OperatorID: cmd.OperatorID,
		Reason:     cmd.Reason,
	})
}
