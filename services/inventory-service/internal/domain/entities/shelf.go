package entities

import (
	"time"
)

type ShelfStatus struct {
	ShelfID       string    `json:"shelf_id"`
	TotalSlots    int       `json:"total_slots"`
	EmptySlots    int       `json:"empty_slots"`
	OccupiedSlots int       `json:"occupied_slots"`
	Slots         []Slot    `json:"slots"`
	UpdatedAt     time.Time `json:"updated_at"`
}