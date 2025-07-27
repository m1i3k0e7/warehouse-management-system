package repositories

import (
	"context"
	"inventory-service/internal/domain/entities"
	"gorm.io/gorm"
)

type SlotRepository interface {
	Create(ctx context.Context, slot *entities.Slot) error
	GetByID(ctx context.Context, id string) (*entities.Slot, error)
	GetByShelfID(ctx context.Context, shelfID string) ([]*entities.Slot, error)
	Update(ctx context.Context, slot *entities.Slot) error
	UpdateWithTx(ctx context.Context, tx *gorm.DB, slot *entities.Slot) error
	BeginTx(ctx context.Context) (*gorm.DB, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Slot, error)
}