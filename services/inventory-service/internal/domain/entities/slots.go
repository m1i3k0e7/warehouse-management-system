package entities

import (
	"time"
)

type SlotStatus string

const (
	SlotStatusEmpty       SlotStatus = "empty"
	SlotStatusOccupied    SlotStatus = "occupied"
	SlotStatusReserved    SlotStatus = "reserved"
	SlotStatusMaintenance SlotStatus = "maintenance"
)

type Slot struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	ShelfID    string     `json:"shelf_id" gorm:"index"`
	Row        int        `json:"row"`
	Column     int        `json:"column"`
	Status     SlotStatus `json:"status"`
	MaterialID *string    `json:"material_id,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Version    int64      `json:"version"` // 樂觀鎖版本號
	
	// 關聯
	Material *Material `json:"material,omitempty" gorm:"foreignKey:MaterialID"`
}

func (Slot) TableName() string {
	return "slots"
}

func (Slot) IsSuitableForMaterialType(materialType string) bool {
	// add logic to determine if the slot is suitable for the given material type
	return true
}