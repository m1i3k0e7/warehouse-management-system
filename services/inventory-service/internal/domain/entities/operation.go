package entities

import (
	"time"
)

type OperationType string

const (
	OperationTypePlacement   OperationType = "placement"
	OperationTypeRemoval     OperationType = "removal"
	OperationTypeMove        OperationType = "move"
	OperationTypeReservation OperationType = "reservation"
)

type OperationStatus string

const (
	OperationStatusPending                   OperationStatus = "pending"
	OperationStatusCompleted                 OperationStatus = "completed"
	OperationStatusFailed                    OperationStatus = "failed"
	OperationStatusCancelled                 OperationStatus = "cancelled"
	OperationStatusPendingPhysicalConfirmation OperationStatus = "pending_physical_confirmation"
)

type Operation struct {
	ID         string          `json:"id" gorm:"primaryKey"`
	Type       OperationType   `json:"type"`
	MaterialID string          `json:"material_id"`
	SlotID     string          `json:"slot_id"`
	OperatorID string          `json:"operator_id"`
	ShelfID    string          `json:"shelf_id"`
	Timestamp  time.Time       `json:"timestamp"`
	Status     OperationStatus `json:"status"`
	
	Material *Material `json:"material,omitempty" gorm:"foreignKey:MaterialID"`
	Slot     *Slot     `json:"slot,omitempty" gorm:"foreignKey:SlotID"`
}

func (Operation) TableName() string {
	return "operations"
}