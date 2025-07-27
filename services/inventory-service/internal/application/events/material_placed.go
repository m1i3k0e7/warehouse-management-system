package events

import (
	"time"

	"github.com/google/uuid"
)

type MaterialPlacedEvent struct {
	EventID    string    `json:"event_id"`
	MaterialID string    `json:"material_id"`
	SlotID     string    `json:"slot_id"`
	ShelfID    string    `json:"shelf_id"`
	OperatorID string    `json:"operator_id"`
	Timestamp  time.Time `json:"timestamp"`
}

func NewMaterialPlacedEvent(materialID, slotID, shelfID, operatorID string) *MaterialPlacedEvent {
	return &MaterialPlacedEvent{
		EventID:    uuid.New().String(),
		MaterialID: materialID,
		SlotID:     slotID,
		ShelfID:    shelfID,
		OperatorID: operatorID,
		Timestamp:  time.Now(),
	}
}
