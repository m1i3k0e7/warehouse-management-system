package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/infrastructure/database/repositories"
)

func TestFailedEventRepository_Create(t *testing.T) {
	repo := repositories.NewFailedEventRepository(db)
	ctx := context.Background()

	event, err := entities.NewFailedEvent(
		"test-failed-event-1",
		"material.placed",
		"material.placed",
		map[string]string{"material_id": "mat123"},
		assert.AnError,
	)
	assert.NoError(t, err)

	err = repo.Create(ctx, event)
	assert.NoError(t, err)

	foundEvent, err := repo.GetByID(ctx, event.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundEvent)
	assert.Equal(t, event.Topic, foundEvent.Topic)
	assert.False(t, foundEvent.Resolved)
}

func TestFailedEventRepository_ListUnresolved(t *testing.T) {
	repo := repositories.NewFailedEventRepository(db)
	ctx := context.Background()

	// Clean up previous test data to ensure isolation
	db.Exec("DELETE FROM failed_events WHERE id LIKE 'test-unresolved-%'")

	event1, _ := entities.NewFailedEvent(
		"test-unresolved-1",
		"shelf.status_changed",
		"shelf.status_changed",
		map[string]string{"shelf_id": "shelf1"},
		assert.AnError,
	)
	event2, _ := entities.NewFailedEvent(
		"test-unresolved-2",
		"system.alert",
		"system.alert",
		map[string]string{"alert_type": "critical"},
		assert.AnError,
	)

	repo.Create(ctx, event1)
	repo.Create(ctx, event2)

	// Create a resolved event to ensure it's not listed
	resolvedEvent, _ := entities.NewFailedEvent(
		"test-resolved-1",
		"material.removed",
		"material.removed",
		map[string]string{"material_id": "mat456"},
		assert.AnError,
	)
	repo.Create(ctx, resolvedEvent)
	repo.MarkAsResolved(ctx, resolvedEvent.ID, "Manually resolved")

	foundEvents, err := repo.ListUnresolved(ctx, 10, 0)
	assert.NoError(t, err)
	assert.True(t, len(foundEvents) >= 2) // May contain events from other tests

	// Check if the newly created unresolved events are in the list
	foundUnresolved1 := false
	foundUnresolved2 := false
	for _, ev := range foundEvents {
		if ev.ID == event1.ID {
			foundUnresolved1 = true
		}
		if ev.ID == event2.ID {
			foundUnresolved2 = true
		}
		assert.False(t, ev.Resolved) // Ensure only unresolved events are returned
	}
	assert.True(t, foundUnresolved1)
	assert.True(t, foundUnresolved2)
}

func TestFailedEventRepository_MarkAsResolved(t *testing.T) {
	repo := repositories.NewFailedEventRepository(db)
	ctx := context.Background()

	event, _ := entities.NewFailedEvent(
		"test-mark-resolved-1",
		"audit.log",
		"audit.log",
		map[string]string{"action": "login"},
		assert.AnError,
	)
	repo.Create(ctx, event)

	notes := "Fixed the Kafka connection issue."
	err := repo.MarkAsResolved(ctx, event.ID, notes)
	assert.NoError(t, err)

	foundEvent, err := repo.GetByID(ctx, event.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundEvent)
	assert.True(t, foundEvent.Resolved)
	assert.NotNil(t, foundEvent.ResolvedAt)
	assert.Equal(t, notes, foundEvent.ResolutionNotes)
}
