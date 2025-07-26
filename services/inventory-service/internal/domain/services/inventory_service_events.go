package services

import (
    "context"
    "time"
    
    "warehouse/internal/domain/entities"
    "warehouse/shared/events"
    "warehouse/pkg/logger"
)

// 發布材料放置事件
func (s *InventoryService) publishMaterialPlacedEvent(ctx context.Context, operation *entities.Operation) {
    event := events.NewMaterialPlacedEvent(
        operation.MaterialID,
        operation.SlotID,
        operation.ShelfID,
        operation.OperatorID,
    )
    
    if err := s.eventService.PublishEvent(ctx, "material.placed", event); err != nil {
        logger.Error("Failed to publish material placed event", err)
        // 發布失敗不影響主流程，但記錄日誌用於後續補償
        s.scheduleEventRetry(ctx, "material.placed", event)
    }
}

// 發布材料移除事件
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

// 發布材料移動事件
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

// 發布料架狀態變更事件
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

// 發布系統告警事件
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


// 事件重試調度器
func (s *InventoryService) scheduleEventRetry(ctx context.Context, topic, eventType string, event interface{}, originalErr error) {
	// 將失敗的事件存儲到數據庫的死信隊列中
	failedEvent, err := entities.NewFailedEvent(generateUUID(), topic, eventType, event, originalErr)
	if err != nil {
		logger.Error("Failed to create failed event", err)
		return
	}

	if err := s.failedEventRepo.Create(ctx, failedEvent); err != nil {
		logger.Error("Failed to save failed event to DLQ", err)
	}
}
