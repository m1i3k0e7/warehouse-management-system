package events

import "time"

// 基礎事件結構
type BaseEvent struct {
    EventID   string    `json:"event_id"`
    EventType string    `json:"event_type"`
    Version   string    `json:"version"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`
}

// 材料放置事件
type MaterialPlacedEvent struct {
    BaseEvent
    MaterialID string `json:"material_id"`
    SlotID     string `json:"slot_id"`
    ShelfID    string `json:"shelf_id"`
    OperatorID string `json:"operator_id"`
}

// 材料移除事件
type MaterialRemovedEvent struct {
    BaseEvent
    MaterialID string `json:"material_id"`
    SlotID     string `json:"slot_id"`
    ShelfID    string `json:"shelf_id"`
    OperatorID string `json:"operator_id"`
}

// 格子狀態變更事件
type SlotStatusChangedEvent struct {
    BaseEvent
    SlotID    string `json:"slot_id"`
    ShelfID   string `json:"shelf_id"`
    OldStatus string `json:"old_status"`
    NewStatus string `json:"new_status"`
    Reason    string `json:"reason"`
}

// 系統告警事件
type SystemAlertEvent struct {
    BaseEvent
    AlertType string                 `json:"alert_type"`
    Severity  string                 `json:"severity"`
    Message   string                 `json:"message"`
    Metadata  map[string]interface{} `json:"metadata"`
}

// 事件工廠
func NewMaterialPlacedEvent(materialID, slotID, shelfID, operatorID string) *MaterialPlacedEvent {
    return &MaterialPlacedEvent{
        BaseEvent: BaseEvent{
            EventID:   generateUUID(),
            EventType: "material.placed",
            Version:   "v1",
            Timestamp: time.Now(),
            Source:    "inventory-service",
        },
        MaterialID: materialID,
        SlotID:     slotID,
        ShelfID:    shelfID,
        OperatorID: operatorID,
    }
}

func NewMaterialRemovedEvent(materialID, slotID, shelfID, operatorID string) *MaterialRemovedEvent {
    return &MaterialRemovedEvent{
        BaseEvent: BaseEvent{
            EventID:   generateUUID(),
            EventType: "material.removed",
            Version:   "v1",
            Timestamp: time.Now(),
            Source:    "inventory-service",
        },
        MaterialID: materialID,
        SlotID:     slotID,
        ShelfID:    shelfID,
        OperatorID: operatorID,
    }
}