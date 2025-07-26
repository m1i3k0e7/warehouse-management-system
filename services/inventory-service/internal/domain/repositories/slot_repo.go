package repositories

import (
	"context"
	"inventory-service/internal/domain/entities"
	"gorm.io/gorm"
)

type MaterialRepository interface {
	Create(ctx context.Context, material *entities.Material) error
	GetByID(ctx context.Context, id string) (*entities.Material, error)
	GetByBarcode(ctx context.Context, barcode string) (*entities.Material, error)
	Update(ctx context.Context, material *entities.Material) error
	UpdateWithTx(ctx context.Context, tx *gorm.DB, material *entities.Material) error
	List(ctx context.Context, limit, offset int) ([]*entities.Material, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Material, error)
}

// File: internal/domain/repositories/slot_repo.go
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