package queries

import (
	"context"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/domain/services"
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
