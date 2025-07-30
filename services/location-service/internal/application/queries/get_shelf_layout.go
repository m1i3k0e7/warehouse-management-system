package queries

import (
	"context"
	"fmt"

	"warehouse/location-service/internal/domain/entities"
	"warehouse/location-service/internal/domain/repositories"
)

type GetShelfLayoutQuery struct {
	ShelfID string `json:"shelf_id"`
}

type GetShelfLayoutQueryHandler struct {
	shelfRepo repositories.ShelfRepository
}

func NewGetShelfLayoutQueryHandler(shelfRepo repositories.ShelfRepository) *GetShelfLayoutQueryHandler {
	return &GetShelfLayoutQueryHandler{
		shelfRepo: shelfRepo,
	}
}

func (h *GetShelfLayoutQueryHandler) Handle(ctx context.Context, query GetShelfLayoutQuery) (*entities.Shelf, error) {
	shelf, err := h.shelfRepo.GetShelfByID(ctx, query.ShelfID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelf layout: %w", err)
	}
	return shelf, nil
}
