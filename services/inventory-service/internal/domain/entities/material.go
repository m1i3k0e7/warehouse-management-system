package entities

import (
	"time"
)

type MaterialStatus string

const (
	MaterialStatusAvailable   MaterialStatus = "available"
	MaterialStatusInUse       MaterialStatus = "in_use"
	MaterialStatusReserved    MaterialStatus = "reserved"
	MaterialStatusMaintenance MaterialStatus = "maintenance"
)

type Material struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	Barcode   string         `json:"barcode" gorm:"uniqueIndex"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Status    MaterialStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (Material) TableName() string {
	return "materials"
}