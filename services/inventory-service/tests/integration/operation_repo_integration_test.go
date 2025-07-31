package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/infrastructure/database/repositories"
)

func TestOperationRepository_Create(t *testing.T) {
	repo := repositories.NewOperationRepository(db)
	ctx := context.Background()

	// Prepare a material and slot for the operation
	material := &entities.Material{
		ID:        "op-mat-1",
		Barcode:   "OP-BARCODE-001",
		Name:      "Operation Test Material",
		Type:      "TEST",
		Status:    entities.MaterialStatusAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(material)

	slot := &entities.Slot{
		ID:      "op-slot-1",
		ShelfID: "OP-SHELF-1",
		Row:     1,
		Column:  1,
		Status:  entities.SlotStatusEmpty,
		UpdatedAt: time.Now(),
		Version: 1,
	}
	db.Create(slot)

	operation := &entities.Operation{
		ID:         "test-operation-1",
		Type:       entities.OperationTypePlacement,
		MaterialID: material.ID,
		SlotID:     slot.ID,
		OperatorID: "operator-1",
		ShelfID:    slot.ShelfID,
		Timestamp:  time.Now(),
		Status:     entities.OperationStatusCompleted,
	}

	err := repo.Create(ctx, operation)
	assert.NoError(t, err)

	foundOperation, err := repo.GetByID(ctx, operation.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundOperation)
	assert.Equal(t, operation.MaterialID, foundOperation.MaterialID)
	assert.Equal(t, operation.SlotID, foundOperation.SlotID)
}

func TestOperationRepository_List(t *testing.T) {
	repo := repositories.NewOperationRepository(db)
	ctx := context.Background()

	// Clean up previous test data to ensure isolation
	db.Exec("DELETE FROM operations WHERE operator_id = ?", "operator-2")

	// Prepare materials and slots for operations
	material1 := &entities.Material{ID: "op-mat-2", Barcode: "OP-BARCODE-002", Name: "Op Mat 2", Type: "TEST", Status: entities.MaterialStatusAvailable, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	material2 := &entities.Material{ID: "op-mat-3", Barcode: "OP-BARCODE-003", Name: "Op Mat 3", Type: "TEST", Status: entities.MaterialStatusAvailable, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(material1)
	db.Create(material2)

	slot1 := &entities.Slot{ID: "op-slot-2", ShelfID: "OP-SHELF-2", Row: 1, Column: 1, Status: entities.SlotStatusEmpty, UpdatedAt: time.Now(), Version: 1}
	slot2 := &entities.Slot{ID: "op-slot-3", ShelfID: "OP-SHELF-2", Row: 1, Column: 2, Status: entities.SlotStatusEmpty, UpdatedAt: time.Now(), Version: 1}
	db.Create(slot1)
	db.Create(slot2)

	operation1 := &entities.Operation{
		ID:         "test-operation-2",
		Type:       entities.OperationTypePlacement,
		MaterialID: material1.ID,
		SlotID:     slot1.ID,
		OperatorID: "operator-2",
		ShelfID:    slot1.ShelfID,
		Timestamp:  time.Now().Add(-1 * time.Hour),
		Status:     entities.OperationStatusCompleted,
	}
	operation2 := &entities.Operation{
		ID:         "test-operation-3",
		Type:       entities.OperationTypeRemoval,
		MaterialID: material2.ID,
		SlotID:     slot2.ID,
		OperatorID: "operator-2",
		ShelfID:    slot2.ShelfID,
		Timestamp:  time.Now(),
		Status:     entities.OperationStatusCompleted,
	}

	repo.Create(ctx, operation1)
	repo.Create(ctx, operation2)

	operations, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.True(t, len(operations) >= 2) // May contain operations from other tests

	// Check if the newly created operations are in the list
	foundOp2 := false
	foundOp3 := false
	for _, op := range operations {
		if op.ID == operation1.ID {
			foundOp2 = true
		}
		if op.ID == operation2.ID {
			foundOp3 = true
		}
	}
	assert.True(t, foundOp2)
	assert.True(t, foundOp3)
}
