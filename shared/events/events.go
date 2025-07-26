package events

import "time"

// 材料移動事件
type MaterialMovedEvent struct {
    BaseEvent
    MaterialID string `json:"material_id"`
    FromSlotID string `json:"from_slot_id"`
    ToSlotID   string `json:"to_slot_id"`
    ShelfID    string `json:"shelf_id"`
    OperatorID string `json:"operator_id"`
}

// 料架狀態變更事件
type ShelfStatusChangedEvent struct {
    BaseEvent
    ShelfID   string `json:"shelf_id"`
    OldStatus string `json:"old_status"`
    NewStatus string `json:"new_status"`
}

// 批量操作事件
type BatchOperationEvent struct {
    BaseEvent
    OperationType string   `json:"operation_type"` // "batch_place", "batch_remove"
    ShelfID       string   `json:"shelf_id"`
    OperatorID    string   `json:"operator_id"`
    ItemCount     int      `json:"item_count"`
    SuccessCount  int      `json:"success_count"`
    FailureCount  int      `json:"failure_count"`
    Duration      int64    `json:"duration_ms"`
}

// 料架健康事件
type ShelfHealthEvent struct {
    BaseEvent
    ShelfID      string  `json:"shelf_id"`
    HealthScore  float64 `json:"health_score"`
    TotalSlots   int     `json:"total_slots"`
    HealthySlots int     `json:"healthy_slots"`
    ErrorSlots   int     `json:"error_slots"`
}

// 事件工廠函數
func NewMaterialMovedEvent(materialID, fromSlotID, toSlotID, shelfID, operatorID string) *MaterialMovedEvent {
    return &MaterialMovedEvent{
        BaseEvent: BaseEvent{
            EventID:   generateUUID(),
            EventType: "material.moved",
            Version:   "v1",
            Timestamp: time.Now(),
            Source:    "inventory-service",
        },
        MaterialID: materialID,
        FromSlotID: fromSlotID,
        ToSlotID:   toSlotID,
        ShelfID:    shelfID,
        OperatorID: operatorID,
    }
}

func NewShelfHealthEvent(shelfID string, healthScore float64, totalSlots, healthySlots, errorSlots int) *ShelfHealthEvent {
    return &ShelfHealthEvent{
        BaseEvent: BaseEvent{
            EventID:   generateUUID(),
            EventType: "shelf.health",
            Version:   "v1",
            Timestamp: time.Now(),
            Source:    "inventory-service",
        },
        ShelfID:      shelfID,
        HealthScore:  healthScore,
        TotalSlots:   totalSlots,
        HealthySlots: healthySlots,
        ErrorSlots:   errorSlots,
    }
}
