package commands

import (
	"context"
	"WMS/services/inventory-service/internal/domain/services"
)

type MoveMaterialCommand struct {
	FromSlotID string
	ToSlotID   string
	OperatorID string
	Reason     string
}

type MoveMaterialCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewMoveMaterialCommandHandler(inventoryService *services.InventoryService) *MoveMaterialCommandHandler {
	return &MoveMaterialCommandHandler{inventoryService: inventoryService}
}

func (h *MoveMaterialCommandHandler) Handle(ctx context.Context, cmd MoveMaterialCommand) error {
	return h.inventoryService.MoveMaterial(ctx, services.MoveMaterialParams{
		FromSlotID: cmd.FromSlotID,
		ToSlotID:   cmd.ToSlotID,
		OperatorID: cmd.OperatorID,
		Reason:     cmd.Reason,
	})
}
