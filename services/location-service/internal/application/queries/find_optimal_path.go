package queries

import (
	"context"

	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/entities"
	"github.com/m1i3k0e7/warehouse-management-system/services/location-service/internal/domain/services"
)

// FindOptimalPathQueryHandler handles the FindOptimalPath query.
	ype FindOptimalPathQueryHandler struct {
	pathfinder *services.PathfindingService
}

// NewFindOptimalPathQueryHandler creates a new FindOptimalPathQueryHandler.
func NewFindOptimalPathQueryHandler(pathfinder *services.PathfindingService) *FindOptimalPathQueryHandler {
	return &FindOptimalPathQueryHandler{pathfinder: pathfinder}
}

// Handle executes the query.
func (h *FindOptimalPathQueryHandler) Handle(ctx context.Context, start, end entities.Point) (*entities.Path, error) {
	return h.pathfinder.FindOptimalPath(start, end)
}