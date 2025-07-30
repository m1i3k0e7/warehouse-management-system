package queries

import (
	"context"
	"fmt"

	"warehouse/location-service/internal/domain/entities"
	"warehouse/location-service/internal/domain/services"
)

type FindOptimalPathQuery struct {
	StartSlotID string `json:"start_slot_id"`
	EndSlotID   string `json:"end_slot_id"`
}

type FindOptimalPathQueryHandler struct {
	pathfindingService *services.PathfindingService
}

func NewFindOptimalPathQueryHandler(pathfindingService *services.PathfindingService) *FindOptimalPathQueryHandler {
	return &FindOptimalPathQueryHandler{
		pathfindingService: pathfindingService,
	}
}

func (h *FindOptimalPathQueryHandler) Handle(ctx context.Context, query FindOptimalPathQuery) (*entities.Path, error) {
	path, err := h.pathfindingService.FindOptimalPath(ctx, query.tStartSlotID, query.EndSlotID)
	if err != nil {
		return nil, fmt.Errorf("failed to find optimal path: %w", err)
	}
	return path, nil
}
