package repositories

import (
	"context"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/domain/repositories"
	
	"gorm.io/gorm"
)

type slotRepository struct {
	db *gorm.DB
}

func NewSlotRepository(db *gorm.DB) repositories.SlotRepository {
	return &slotRepository{db: db}
}

func (r *slotRepository) Create(ctx context.Context, slot *entities.Slot) error {
	return r.db.WithContext(ctx).Create(slot).Error
}

func (r *slotRepository) GetByID(ctx context.Context, id string) (*entities.Slot, error) {
	var slot entities.Slot
	err := r.db.WithContext(ctx).
		Preload("Material").
		Where("id = ?", id).
		First(&slot).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *slotRepository) GetByShelfID(ctx context.Context, shelfID string) ([]*entities.Slot, error) {
	var slots []*entities.Slot
	err := r.db.WithContext(ctx).
		Preload("Material").
		Where("shelf_id = ?", shelfID).
		Order("row, column").
		Find(&slots).Error
	return slots, err
}

func (r *slotRepository) Update(ctx context.Context, slot *entities.Slot) error {
	return r.db.WithContext(ctx).Save(slot).Error
}

func (r *slotRepository) UpdateWithTx(ctx context.Context, tx *gorm.DB, slot *entities.Slot) error {
	return tx.WithContext(ctx).
		Where("version = ?", slot.Version-1).
		Save(slot).Error
}

func (r *slotRepository) BeginTx(ctx context.Context) (*gorm.DB, error) {
	return r.db.WithContext(ctx).Begin(), nil
}

func (r *slotRepository) GetEmptySlotsByShelf(ctx context.Context, shelfID string) ([]*entities.Slot, error) {
	var slots []*entities.Slot
	err := r.db.WithContext(ctx).
		Where("shelf_id = ? AND status = ?", shelfID, entities.SlotStatusEmpty).
		Order("row, column").
		Find(&slots).Error
	return slots, err
}

func (r *slotRepository) List(ctx context.Context, limit, offset int) ([]*entities.Slot, error) {
	var slots []*entities.Slot
	err := r.db.WithContext(ctx).
		Preload("Material").
		Limit(limit).
		Offset(offset).
		Order("shelf_id, row, column").
		Find(&slots).Error
	return slots, err
}