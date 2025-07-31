package commands

import (
	"context"
	"WMS/services/inventory-service/internal/domain/services"
)

type ReserveSlotsCommand struct {
	SlotIDs    []string
	OperatorID string
	Duration   int
	Purpose    string
}

type ReserveSlotsCommandHandler struct {
	inventoryService *services.InventoryService
}

func NewReserveSlotsCommandHandler(inventoryService *services.InventoryService) *ReserveSlotsCommandHandler {
	return &ReserveSlotsCommandHandler{inventoryService: inventoryService}
}

func (h *ReserveSlotsCommandHandler) Handle(ctx context.Context, cmd ReserveSlotsCommand) error {
	return h.inventoryService.ReserveSlots(ctx, services.ReserveSlotsParams{
		SlotIDs:    cmd.SlotIDs,
		OperatorID: cmd.OperatorID,
		Duration:   cmd.Duration,
		Purpose:    cmd.Purpose,
	})
}
