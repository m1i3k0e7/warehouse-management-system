package commands

import (
	"context"
	"warehouse/internal/domain/services"
)

type HandleSlotErrorCommand struct {
	SlotID    string
	ErrorType string
}

type HandleSlotErrorCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewHandleSlotErrorCommandHandler(inventoryService *services.InventoryService) *HandleSlotErrorCommandHandler {
	return &HandleSlotErrorCommandHandler{inventoryService: inventoryService}
}

func (h *HandleSlotErrorCommandHandler) Handle(ctx context.Context, cmd HandleSlotErrorCommand) error {
	return h.inventoryService.HandleSlotError(ctx, cmd.SlotID, cmd.ErrorType)
}
