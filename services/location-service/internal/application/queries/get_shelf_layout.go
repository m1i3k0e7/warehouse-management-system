package queries

import (
	"context"

	"github.com/your-repo/wms/location-service/internal/domain/entities"
	"github.com/your-repo/wms/location-service/internal/domain/repositories"
)

// GetShelfLayoutQueryHandler handles the GetShelfLayout query.
	ype GetShelfLayoutQueryHandler struct {
	shelfRepo repositories.ShelfRepository
}

// NewGetShelfLayoutQueryHandler creates a new GetShelfLayoutQueryHandler.
func NewGetShelfLayoutQueryHandler(shelfRepo repositories.ShelfRepository) *GetShelfLayoutQueryHandler {
	return &GetShelfLayoutQueryHandler{shelfRepo: shelfRepo}
}

// Handle executes the query.
func (h *GetShelfLayoutQueryHandler) Handle(ctx context.Context, shelfID string) (*entities.Shelf, error) {
	return h.shelfRepo.FindByID(ctx, shelfID)
}