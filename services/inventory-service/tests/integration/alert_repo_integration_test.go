package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"warehouse/internal/domain/entities"
	"warehouse/internal/infrastructure/database/repositories"
)

func TestAlertRepository_Create(t *testing.T) {
	repo := repositories.NewAlertRepository(db)
	ctx := context.Background()

	alert := &entities.Alert{
		ID:        "test-alert-1",
		Type:      "shelf_health",
		ShelfID:   "SHELF-001",
		Message:   "Shelf health critical",
		Severity:  "critical",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, alert)
	assert.NoError(t, err)

	foundAlert, err := repo.GetByID(ctx, alert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundAlert)
	assert.Equal(t, alert.Message, foundAlert.Message)
}

func TestAlertRepository_UpdateStatus(t *testing.T) {
	repo := repositories.NewAlertRepository(db)
	ctx := context.Background()

	alert := &entities.Alert{
		ID:        "test-alert-2",
		Type:      "slot_error",
		SlotID:    "SLOT-001",
		Message:   "Slot sensor malfunction",
		Severity:  "high",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, alert)
	assert.NoError(t, err)

	err = repo.UpdateStatus(ctx, alert.ID, "acknowledged")
	assert.NoError(t, err)

	foundAlert, err := repo.GetByID(ctx, alert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundAlert)
	assert.Equal(t, "acknowledged", foundAlert.Status)
}

func TestAlertRepository_MarkAsResolved(t *testing.T) {
	repo := repositories.NewAlertRepository(db)
	ctx := context.Background()

	alert := &entities.Alert{
		ID:        "test-alert-3",
		Type:      "system_warning",
		Message:   "System disk space low",
		Severity:  "medium",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, alert)
	assert.NoError(t, err)

	err = repo.MarkAsResolved(ctx, alert.ID)
	assert.NoError(t, err)

	foundAlert, err := repo.GetByID(ctx, alert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundAlert)
	assert.Equal(t, "resolved", foundAlert.Status)
	assert.NotNil(t, foundAlert.ResolvedAt)
}
