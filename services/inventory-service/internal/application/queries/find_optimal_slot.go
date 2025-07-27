package queries

import (
	"context"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/services"
)

type FindOptimalSlotQuery struct {
	MaterialType string
	ShelfID      string
}

type FindOptimalSlotQueryHandler struct {
	inventoryService *services.InventoryService
}

func NewFindOptimalSlotQueryHandler(inventoryService *services.InventoryService) *FindOptimalSlotQueryHandler {
	return &FindOptimalSlotQueryHandler{inventoryService: inventoryService}
}

func (h *FindOptimalSlotQueryHandler) Handle(ctx context.Context, query FindOptimalSlotQuery) (*entities.Slot, error) {
	return h.inventoryService.FindOptimalSlot(ctx, query.MaterialType, query.ShelfID)
}
