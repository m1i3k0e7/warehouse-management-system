package queries

import (
	"context"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/domain/repositories"
)

type GetOperationsQuery struct {
	Limit  int
	Offset int
}

type GetOperationsQueryHandler struct {
	operationRepo repositories.OperationRepository
}

func NewGetOperationsQueryHandler(operationRepo repositories.OperationRepository) *GetOperationsQueryHandler {
	return &GetOperationsQueryHandler{operationRepo: operationRepo}
}

func (h *GetOperationsQueryHandler) Handle(ctx context.Context, query GetOperationsQuery) ([]*entities.Operation, error) {
	return h.operationRepo.List(ctx, query.Limit, query.Offset)
}
