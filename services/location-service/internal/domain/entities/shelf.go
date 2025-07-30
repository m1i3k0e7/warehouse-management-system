package entities

import (
	"time"
)

type Shelf struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Zone      string    `json:"zone" bson:"zone"`
	Rows      int32     `json:"rows" bson:"rows"`
	Columns   int32     `json:"columns" bson:"columns"`
	Levels    int32     `json:"levels" bson:"levels"`
	Slots     []Slot    `json:"slots" bson:"slots"` // Embedded slots
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type Slot struct {
	ID          string `json:"id" bson:"id"`
	ShelfID     string `json:"shelf_id" bson:"shelf_id"`
	Row         int32  `json:"row" bson:"row"`
	Column      int32  `json:"column" bson:"column"`
	Level       int32  `json:"level" bson:"level"`
	Status      string `json:"status" bson:"status"` // e.g., "empty", "occupied", "reserved", "maintenance"
	MaterialID  string `json:"material_id,omitempty" bson:"material_id,omitempty"`
	// Add more properties like capacity, material type restrictions, etc.
}
