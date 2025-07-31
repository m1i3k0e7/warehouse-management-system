package repositories

import (
	"context"
	"time"

	"WMS/services/inventory-service/internal/domain/entities"
	"gorm.io/gorm"
)

type OperationRepository interface {
	Create(ctx context.Context, operation *entities.Operation) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, operation *entities.Operation) error
	GetByID(ctx context.Context, id string) (*entities.Operation, error)
	GetByShelfID(ctx context.Context, shelfID string, limit, offset int) ([]*entities.Operation, error)
	GetByOperatorID(ctx context.Context, operatorID string, limit, offset int) ([]*entities.Operation, error)
	List(ctx context.Context, limit int, offset int) ([]*entities.Operation, error)
	GetTimedOutPendingPhysicalConfirmations(ctx context.Context, timeout time.Duration) ([]*entities.Operation, error)
	GetPendingPhysicalConfirmationsBySlotID(ctx context.Context, slotID string) ([]*entities.Operation, error)
	GetPendingRemovalConfirmationsBySlotID(ctx context.Context, slotID string) ([]*entities.Operation, error)
	BeginTx(ctx context.Context) (*gorm.DB, error)
	UpdateWithTx(ctx context.Context, tx *gorm.DB, operation *entities.Operation) error
}
