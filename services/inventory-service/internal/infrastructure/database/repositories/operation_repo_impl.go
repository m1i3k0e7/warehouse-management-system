package repositories

import (
	"context"
	"inventory-service/internal/domain/entities"
	"inventory-service/internal/domain/repositories"

	"gorm.io/gorm"
)

type operationRepository struct {
	db *gorm.DB
}

func NewOperationRepository(db *gorm.DB) repositories.OperationRepository {
	return &operationRepository{db: db}
}

func (r *operationRepository) Create(ctx context.Context, operation *entities.Operation) error {
	return r.db.WithContext(ctx).Create(operation).Error
}

func (r *operationRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, operation *entities.Operation) error {
	return tx.WithContext(ctx).Create(operation).Error
}

func (r *operationRepository) GetByID(ctx context.Context, id string) (*entities.Operation, error) {
	var operation entities.Operation
	err := r.db.WithContext(ctx).Preload("Material").Preload("Slot").First(&operation, "id = ?", id).Error
	return &operation, err
}

func (r *operationRepository) GetByShelfID(ctx context.Context, shelfID string, limit, offset int) ([]*entities.Operation, error) {
	var operations []*entities.Operation
	err := r.db.WithContext(ctx).
		Where("shelf_id = ?", shelfID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&operations).Error
	return operations, err
}

func (r *operationRepository) GetByOperatorID(ctx context.Context, operatorID string, limit, offset int) ([]*entities.Operation, error) {
	var operations []*entities.Operation
	err := r.db.WithContext(ctx).
		Where("operator_id = ?", operatorID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Preload("Material").
		Preload("Slot").
		Find(&operations).Error
	return operations, err
}

func (r *operationRepository) List(ctx context.Context, limit, offset int) ([]*entities.Operation, error) {
	var operations []*entities.Operation
	err := r.db.WithContext(ctx).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Preload("Material").
		Preload("Slot").
		Find(&operations).Error
	return operations, err
}

func (r *operationRepository) GetTimedOutPendingPhysicalConfirmations(ctx context.Context, timeout time.Duration) ([]*entities.Operation, error) {
	var operations []*entities.Operation
	err := r.db.WithContext(ctx).
		Where("status = ? AND timestamp < ?", entities.OperationStatusPendingPhysicalConfirmation, time.Now().Add(-timeout)).
		Find(&operations).Error
	return operations, err
}
