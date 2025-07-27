package queries

import (
	"context"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/services"
)

type SearchMaterialsQuery struct {
	Query  string
	Limit  int
	Offset int
}

type SearchMaterialsQueryHandler struct {
	inventoryService *services.InventoryService
}

func NewSearchMaterialsQueryHandler(inventoryService *services.InventoryService) *SearchMaterialsQueryHandler {
	return &SearchMaterialsQueryHandler{inventoryService: inventoryService}
}

func (h *SearchMaterialsQueryHandler) Handle(ctx context.Context, query SearchMaterialsQuery) ([]*entities.Material, error) {
	return h.inventoryService.SearchMaterials(ctx, query.Query, query.Limit, query.Offset)
}
