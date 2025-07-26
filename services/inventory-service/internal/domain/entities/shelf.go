package entities

import (
	"time"
)

type ShelfStatus struct {
	ShelfID       string    `json:"shelf_id"`
	TotalSlots    int       `json:"total_slots"`
	EmptySlots    int       `json:"empty_slots"`
	OccupiedSlots int       `json:"occupied_slots"`
	Slots         []Slot    `json:"slots"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// File: internal/domain/repositories/material_repo.go
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