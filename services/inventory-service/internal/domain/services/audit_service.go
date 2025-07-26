package services

import (
    "context"
    "encoding/json"
    "time"
    
    "warehouse/internal/domain/entities"
    "warehouse/pkg/logger"
)

type AuditService struct {
    eventService *EventService
}

func NewAuditService(eventService *EventService) *AuditService {
    return &AuditService{eventService: eventService}
}

type AuditLog struct {
    ID          string                 `json:"id"`
    Action      string                 `json:"action"`
    EntityType  string                 `json:"entity_type"`
    EntityID    string                 `json:"entity_id"`
    OperatorID  string                 `json:"operator_id"`
    Changes     map[string]interface{} `json:"changes,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Success     bool                   `json:"success"`
    ErrorMsg    string                 `json:"error_message,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
}

func (s *AuditService) LogSuccessfulOperation(ctx context.Context, operation *entities.Operation) {
    auditLog := &AuditLog{
        ID:         generateUUID(),
        Action:     string(operation.Type),
        EntityType: "material",
        EntityID:   operation.MaterialID,
        OperatorID: operation.OperatorID,
        Metadata: map[string]interface{}{
            "slot_id":  operation.SlotID,
            "shelf_id": operation.ShelfID,
        },
        Success:   true,
        Timestamp: time.Now(),
    }
    
    s.publishAuditLog(ctx, auditLog)
}

func (s *AuditService) LogFailedOperation(ctx context.Context, action string, command interface{}, err error) {
    auditLog := &AuditLog{
        ID:        generateUUID(),
        Action:    action,
        Success:   false,
        ErrorMsg:  err.Error(),
        Metadata:  map[string]interface{}{"command": command},
        Timestamp: time.Now(),
    }
    
    s.publishAuditLog(ctx, auditLog)
}

func (s *AuditService) publishAuditLog(ctx context.Context, log *AuditLog) {
    if err := s.eventService.PublishEvent(ctx, "audit.log", log); err != nil {
        logger.Error("Failed to publish audit log", err)
    }
}