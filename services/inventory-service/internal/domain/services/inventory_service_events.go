package services

import (
	"context"
	"time"

	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/pkg/utils/logger"
)

func (s *InventoryService) SaveFailedEventToDLQ(ctx context.Context, topic, eventType string, event any, originalErr error) {
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
		s.SaveFailedEventToDLQ(ctx, EventTypeShelfStatusChanged, EventTypeShelfStatusChanged, event, err)
	}
}

func (s *InventoryService) publishPhysicalPlacementRequestedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		OperationID string    `json:"operation_id"`
		MaterialID  string    `json:"material_id"`
		SlotID      string    `json:"slot_id"`
		ShelfID     string    `json:"shelf_id"`
		OperatorID  string    `json:"operator_id"`
		Timestamp   time.Time `json:"timestamp"`
		EventType   string    `json:"event_type"`
	}{
		OperationID: operation.ID,
		MaterialID:  operation.MaterialID,
		SlotID:      operation.SlotID,
		ShelfID:     operation.ShelfID,
		OperatorID:  operation.OperatorID,
		Timestamp:   time.Now(),
		EventType:   EventTypePhysicalPlacementRequested,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypePhysicalPlacementRequested, event); err != nil {
		logger.Error("Failed to publish physical placement requested event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypePhysicalPlacementRequested, EventTypePhysicalPlacementRequested, event, err)
	}
}

func (s *InventoryService) publishPhysicalPlacementConfirmedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		OperationID string    `json:"operation_id"`
		MaterialID  string    `json:"material_id"`
		SlotID      string    `json:"slot_id"`
		ShelfID     string    `json:"shelf_id"`
		OperatorID  string    `json:"operator_id"`
		Timestamp   time.Time `json:"timestamp"`
		EventType   string    `json:"event_type"`
	}{
		OperationID: operation.ID,
		MaterialID:  operation.MaterialID,
		SlotID:      operation.SlotID,
		ShelfID:     operation.ShelfID,
		OperatorID:  operation.OperatorID,
		Timestamp:   time.Now(),
		EventType:   EventTypePhysicalPlacementConfirmed,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypePhysicalPlacementConfirmed, event); err != nil {
		logger.Error("Failed to publish physical placement confirmed event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypePhysicalPlacementConfirmed, EventTypePhysicalPlacementConfirmed, event, err)
	}
}

func (s *InventoryService) publishPhysicalPlacementFailedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		OperationID string    `json:"operation_id"`
		MaterialID  string    `json:"material_id"`
		SlotID      string    `json:"slot_id"`
		ShelfID     string    `json:"shelf_id"`
		OperatorID  string    `json:"operator_id"`
		Timestamp   time.Time `json:"timestamp"`
		EventType   string    `json:"event_type"`
	}{
		OperationID: operation.ID,
		MaterialID:  operation.MaterialID,
		SlotID:      operation.SlotID,
		ShelfID:     operation.ShelfID,
		OperatorID:  operation.OperatorID,
		Timestamp:   time.Now(),
		EventType:   EventTypePhysicalPlacementFailed,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypePhysicalPlacementFailed, event); err != nil {
		logger.Error("Failed to publish physical placement failed event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypePhysicalPlacementFailed, EventTypePhysicalPlacementFailed, event, err)
	}
}

func (s *InventoryService) publishPhysicalRemovalConfirmedEvent(ctx context.Context, operation *entities.Operation) {}

func (s *InventoryService) publishPhysicalRemovalFailedEvent(ctx context.Context, operation *entities.Operation) {}

func (s *InventoryService) publishUnplannedPlacementEvent(ctx context.Context, slotID, materialBarcode string) {
	event := struct {
		SlotID          string    `json:"slot_id"`
		MaterialBarcode string    `json:"material_barcode"`
		Timestamp       time.Time `json:"timestamp"`
		EventType       string    `json:"event_type"`
	}{
		SlotID:          slotID,
		MaterialBarcode: materialBarcode,
		Timestamp:       time.Now(),
		EventType:       EventTypeUnplannedPlacement,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeUnplannedPlacement, event); err != nil {
		logger.Error("Failed to publish unplanned placement event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypeUnplannedPlacement, EventTypeUnplannedPlacement, event, err)
	}
}

func (s *InventoryService) publishUnplannedRemovalEvent(ctx context.Context, slotID string, materialBarcode string) {
	event := struct {
		SlotID          string    `json:"slot_id"`
		MaterialBarcode string    `json:"material_barcode"`
		Timestamp       time.Time `json:"timestamp"`
		EventType       string    `json:"event_type"`
	}{
		SlotID:          slotID,
		MaterialBarcode: materialBarcode,
		Timestamp:       time.Now(),
		EventType:       EventTypeUnplannedRemoval,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeUnplannedRemoval, event); err != nil {
		logger.Error("Failed to publish unplanned removal event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypeUnplannedRemoval, EventTypeUnplannedRemoval, event, err)
	}
}

func (s *InventoryService) publishMaterialPlacedEvent(ctx context.Context, operation *entities.Operation) {
	event := struct {
		EventID    string    `json:"event_id"`
		MaterialID string    `json:"material_id"`
		SlotID     string    `json:"slot_id"`
		ShelfID    string    `json:"shelf_id"`
		OperatorID string    `json:"operator_id"`
		Timestamp  time.Time `json:"timestamp"`
		EventType  string    `json:"event_type"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		SlotID:     operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
		EventType:  EventTypeMaterialPlaced,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialPlaced, event); err != nil {
		logger.Error("Failed to publish material placed event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypeMaterialPlaced, EventTypeMaterialPlaced, event, err)
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
		EventType  string    `json:"event_type"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		SlotID:     operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
		EventType:  EventTypeMaterialRemoved,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialRemoved, event); err != nil {
		logger.Error("Failed to publish material removed event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypeMaterialRemoved, EventTypeMaterialRemoved, event, err)
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
		EventType  string    `json:"event_type"`
	}{
		EventID:    generateUUID(),
		MaterialID: operation.MaterialID,
		FromSlotID: fromSlotID,
		ToSlotID:   operation.SlotID,
		ShelfID:    operation.ShelfID,
		OperatorID: operation.OperatorID,
		Timestamp:  time.Now(),
		EventType:  EventTypeMaterialMoved,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeMaterialMoved, event); err != nil {
		logger.Error("Failed to publish material moved event", err)
		s.SaveFailedEventToDLQ(ctx, EventTypeMaterialMoved, EventTypeMaterialMoved, event, err)
	}
}
