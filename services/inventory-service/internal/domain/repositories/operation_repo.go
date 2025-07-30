package repositories

import (
	"context"
	"time"

	"inventory-service/internal/domain/entities"
	"gorm.io/gorm"
)

type OperationRepository interface {
	Create(ctx context.Context, operation *entities.Operation) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, operation *entities.Operation) error
	GetByID(ctx context.Context, id string) (*entities.Operation, error)
	GetByShelfID(ctx context.Context, shelfID string, limit, offset int) ([]*entities.Operation, error)
	GetByOperatorID(ctx context.Context, operatorID string, limit, offset int) ([]*entities.Operation, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Operation, error)
	GetTimedOutPendingPhysicalConfirmations(ctx context.Context, timeout time.Duration) ([]*entities.Operation, error)
}