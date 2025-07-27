package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"warehouse/internal/domain/entities"
	"warehouse/internal/infrastructure/database/repositories"
)

func TestSlotRepository_Create(t *testing.T) {
	repo := repositories.NewSlotRepository(db)
	ctx := context.Background()

	slot := &entities.Slot{
		ID:      "test-slot-1",
		ShelfID: "test-shelf-1",
		Row:     1,
		Column:  1,
		Status:  entities.SlotStatusEmpty,
		UpdatedAt: time.Now(),
		Version: 1,
	}

	err := repo.Create(ctx, slot)
	assert.NoError(t, err)

	foundSlot, err := repo.GetByID(ctx, slot.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundSlot)
	assert.Equal(t, slot.ShelfID, foundSlot.ShelfID)
}

func TestSlotRepository_GetByShelfID(t *testing.T) {
	repo := repositories.NewSlotRepository(db)
	ctx := context.Background()

	// Clean up previous test data to ensure isolation
	db.Exec("DELETE FROM slots WHERE shelf_id = ?", "test-shelf-2")

	slot1 := &entities.Slot{
		ID:      "test-slot-2-1",
		ShelfID: "test-shelf-2",
		Row:     1,
		Column:  1,
		Status:  entities.SlotStatusEmpty,
		UpdatedAt: time.Now(),
		Version: 1,
	}
	slot2 := &entities.Slot{
		ID:      "test-slot-2-2",
		ShelfID: "test-shelf-2",
		Row:     1,
		Column:  2,
		Status:  entities.SlotStatusOccupied,
		UpdatedAt: time.Now(),
		Version: 1,
	}

	err := repo.Create(ctx, slot1)
	assert.NoError(t, err)
	err = repo.Create(ctx, slot2)
	assert.NoError(t, err)

	foundSlots, err := repo.GetByShelfID(ctx, "test-shelf-2")
	assert.NoError(t, err)
	assert.Len(t, foundSlots, 2)
	assert.Equal(t, "test-slot-2-1", foundSlots[0].ID)
	assert.Equal(t, "test-slot-2-2", foundSlots[1].ID)
}

func TestSlotRepository_UpdateWithTx(t *testing.T) {
	repo := repositories.NewSlotRepository(db)
	ctx := context.Background()

	slot := &entities.Slot{
		ID:      "test-slot-3",
		ShelfID: "test-shelf-3",
		Row:     1,
		Column:  1,
		Status:  entities.SlotStatusEmpty,
		UpdatedAt: time.Now(),
		Version: 1,
	}
	err := repo.Create(ctx, slot)
	assert.NoError(t, err)

	tx := db.Begin()
	assert.NotNil(t, tx)

	slot.Status = entities.SlotStatusOccupied
	slot.Version++ // Increment version for optimistic locking

	err = repo.UpdateWithTx(ctx, tx, slot)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	foundSlot, err := repo.GetByID(ctx, slot.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundSlot)
	assert.Equal(t, entities.SlotStatusOccupied, foundSlot.Status)
	assert.Equal(t, int64(2), foundSlot.Version)
}

func TestSlotRepository_UpdateWithTx_OptimisticLocking(t *testing.T) {
	repo := repositories.NewSlotRepository(db)
	ctx := context.Background()

	slot := &entities.Slot{
		ID:      "test-slot-4",
		ShelfID: "test-shelf-4",
		Row:     1,
		Column:  1,
		Status:  entities.SlotStatusEmpty,
		UpdatedAt: time.Now(),
		Version: 1,
	}
	err := repo.Create(ctx, slot)
	assert.NoError(t, err)

	// Simulate another transaction updating the slot
	anotherTx := db.Begin()
	anotherSlot := &entities.Slot{}
	anotherTx.First(anotherSlot, "id = ?", slot.ID)
	anotherSlot.Status = entities.SlotStatusReserved
	anotherSlot.Version++
	anotherTx.Save(anotherSlot)
	anotherTx.Commit()

	// Now, try to update with the old version
	currentTx := db.Begin()
	slot.Status = entities.SlotStatusOccupied
	slot.Version++ // This will make slot.Version-1 equal to the original version, not the updated one

	err = repo.UpdateWithTx(ctx, currentTx, slot)
	assert.Error(t, err) // Expect an error due to optimistic locking
	assert.Contains(t, err.Error(), "no rows affected") // GORM returns this for optimistic locking failure

	currentTx.Rollback()
}
