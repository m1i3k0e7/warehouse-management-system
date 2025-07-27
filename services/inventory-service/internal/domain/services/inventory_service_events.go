package services

import (
	"context"
	"time"

	"warehouse/internal/domain/entities"
	"warehouse/pkg/logger"
)

func (s *InventoryService) publishMaterialPlacedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		EventID    string    `json:"event_id"`
		MaterialID string    `json:"material_id"`
		SlotID     string    `json:"slot_id"`
		ShelfID    string    `json:"shelf_id"`
		OperatorID string    `json:"operator_id"`
		Timestamp  time.Time `json:"timestamp"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		SlotID:     operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialPlaced, event); err != nil {
		logger.Error("Failed to publish material placed event", err)
		s.scheduleEventRetry(ctx, EventTypeMaterialPlaced, EventTypeMaterialPlaced, event, err)
	}
}

func (s *InventoryService) publishMaterialRemovedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		EventID    string    `json:"event_id"`
		MaterialID string    `json:"material_id"`
		SlotID     string    `json:"slot_id"`
		ShelfID    string    `json:"shelf_id"`
		OperatorID string    `json:"operator_id"`
		Timestamp  time.Time `json:"timestamp"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		SlotID:     operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialRemoved, event); err != nil {
		logger.Error("Failed to publish material removed event", err)
		s.scheduleEventRetry(ctx, EventTypeMaterialRemoved, EventTypeMaterialRemoved, event, err)
	}
}

func (s *InventoryService) publishMaterialMovedEvent(ctx context.Context, operation *entities.Operation, fromSlotID string) {
	event := struct {
		EventID    string    `json:"event_id"`
		MaterialID string    `json:"material_id"`
		FromSlotID string    `json:"from_slot_id"`
		ToSlotID   string    `json:"to_slot_id"`
		ShelfID    string    `json:"shelf_id"`
		OperatorID string    `json:"operator_id"`
		Timestamp  time.Time `json:"timestamp"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		FromSlotID: fromSlotID,
		ToSlotID:   operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialMoved, event); err != nil {
		logger.Error("Failed to publish material moved event", err)
		s.scheduleEventRetry(ctx, EventTypeMaterialMoved, EventTypeMaterialMoved, event, err)
	}
}

func (s *InventoryService) publishShelfStatusChangedEvent(ctx context.Context, shelfID string, oldStatus, newStatus string) {
	event := struct {
		EventID   string    `json:"event_id"`
		ShelfID   string    `json:"shelf_id"`
		OldStatus string    `json:"old_status"`
		NewStatus string    `json:"new_status"`
		Timestamp time.Time `json:"timestamp"`
	}{
		EventID:   generateUUID(),
		ShelfID:   shelfID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Timestamp: time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeShelfStatusChanged, event); err != nil {
		logger.Error("Failed to publish shelf status changed event", err)
		s.scheduleEventRetry(ctx, EventTypeShelfStatusChanged, EventTypeShelfStatusChanged, event, err)
	}
}

func (s *InventoryService) publishSystemAlertEvent(ctx context.Context, alertType, severity, message string, metadata map[string]interface{}) {
	event := struct {
		EventID   string                 `json:"event_id"`
		AlertType string                 `json:"alert_type"`
		Severity  string                 `json:"severity"`
		Message   string                 `json:"message"`
		Metadata  map[string]interface{} `json:"metadata"`
		Timestamp time.Time              `json:"timestamp"`
	}{
		EventID:   generateUUID(),
		AlertType: alertType,
		Severity:  severity,
		Message:   message,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeSystemAlert, event); err != nil {
		logger.Error("Failed to publish system alert event", err)
		s.scheduleEventRetry(ctx, EventTypeSystemAlert, EventTypeSystemAlert, event, err)
	}
}

func (s *InventoryService) scheduleEventRetry(ctx context.Context, topic, eventType string, event interface{}, originalErr error) {
	// save the failed event to the dead-letter queue (DLQ)
	failedEvent, err := entities.NewFailedEvent(generateUUID(), topic, eventType, event, originalErr)
	if err != nil {
		logger.Error("Failed to create failed event", err)
		return
	}

	if err := s.failedEventRepo.Create(ctx, failedEvent); err != nil {
		logger.Error("Failed to save failed event to DLQ", err)
	}
}