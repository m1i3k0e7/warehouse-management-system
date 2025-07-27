package services

import (
    "context"
    "time"
    
    "warehouse/internal/domain/entities"
    "warehouse/shared/events"
    "warehouse/pkg/logger"
)

func (s *InventoryService) publishMaterialPlacedEvent(ctx context.Context, operation *entities.Operation) {
    event := events.NewMaterialPlacedEvent(
        operation.MaterialID,
        operation.SlotID,
        operation.ShelfID,
        operation.OperatorID,
    )
    
    if err := s.eventService.PublishEvent(ctx, "material.placed", event); err != nil {
        logger.Error("Failed to publish material placed event", err)
        s.scheduleEventRetry(ctx, "material.placed", event)
    }
}

func (s *InventoryService) publishMaterialRemovedEvent(ctx context.Context, operation *entities.Operation) {
    event := events.NewMaterialRemovedEvent(
        operation.MaterialID,
        operation.SlotID,
        operation.ShelfID,
        operation.OperatorID,
    )
    
    if err := s.eventService.PublishEvent(ctx, "material.removed", event); err != nil {
        logger.Error("Failed to publish material removed event", err)
        s.scheduleEventRetry(ctx, "material.removed", event)
    }
}

func (s *InventoryService) publishMaterialMovedEvent(ctx context.Context, operation *entities.Operation, fromSlotID string) {
    event := &events.MaterialMovedEvent{
        BaseEvent: events.BaseEvent{
            EventID:   generateUUID(),
            EventType: "material.moved",
            Version:   "v1",
            Timestamp: time.Now(),
            Source:    "inventory-service",
        },
        MaterialID: operation.MaterialID,
        FromSlotID: fromSlotID,
        ToSlotID:   operation.SlotID,
        ShelfID:    operation.ShelfID,
        OperatorID: operation.OperatorID,
    }
    
    if err := s.eventService.PublishEvent(ctx, "material.moved", event); err != nil {
        logger.Error("Failed to publish material moved event", err)
        s.scheduleEventRetry(ctx, "material.moved", event)
    }
}

func (s *InventoryService) publishShelfStatusChangedEvent(ctx context.Context, shelfID string, oldStatus, newStatus string) {
    event := &events.ShelfStatusChangedEvent{
        BaseEvent: events.BaseEvent{
            EventID:   generateUUID(),
            EventType: "shelf.status_changed",
            Version:   "v1",
            Timestamp: time.Now(),
            Source:    "inventory-service",
        },
        ShelfID:   shelfID,
        OldStatus: oldStatus,
        NewStatus: newStatus,
    }
    
    if err := s.eventService.PublishEvent(ctx, "shelf.status_changed", event); err != nil {
        logger.Error("Failed to publish shelf status changed event", err)
    }
}

func (s *InventoryService) publishSystemAlertEvent(ctx context.Context, alertType, severity, message string, metadata map[string]interface{}) {
    event := &events.SystemAlertEvent{
        BaseEvent: events.BaseEvent{
            EventID:   generateUUID(),
            EventType: "system.alert",
            Version:   "v1",
            Timestamp: time.Now(),
            Source:    "inventory-service",
        },
        AlertType: alertType,
        Severity:  severity,
        Message:   message,
        Metadata:  metadata,
    }
    
    if err := s.eventService.PublishEvent(ctx, "system.alert", event); err != nil {
        logger.Error("Failed to publish system alert event", err)
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
