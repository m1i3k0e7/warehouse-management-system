package queries

import (
	"context"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/domain/services"
)

type GetShelfStatusQuery struct {
	ShelfID string
}

type GetShelfStatusQueryHandler struct {
	inventoryService *services.InventoryService
}

func NewGetShelfStatusQueryHandler(inventoryService *services.InventoryService) *GetShelfStatusQueryHandler {
	return &GetShelfStatusQueryHandler{inventoryService: inventoryService}
}

func (h *GetShelfStatusQueryHandler) Handle(ctx context.Context, query GetShelfStatusQuery) (*entities.ShelfStatus, error) {
	return h.inventoryService.GetShelfStatus(ctx, query.ShelfID)
}
