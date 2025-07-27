package queries

import (
	"context"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/services"
)

type HealthCheckShelfQuery struct {
	ShelfID string
}

type HealthCheckShelfQueryHandler struct {
	inventoryService *services.InventoryService
}

func NewHealthCheckShelfQueryHandler(inventoryService *services.InventoryService) *HealthCheckShelfQueryHandler {
	return &HealthCheckShelfQueryHandler{inventoryService: inventoryService}
}

func (h *HealthCheckShelfQueryHandler) Handle(ctx context.Context, query HealthCheckShelfQuery) (*entities.ShelfHealth, error) {
	return h.inventoryService.HealthCheckShelf(ctx, query.ShelfID)
}
