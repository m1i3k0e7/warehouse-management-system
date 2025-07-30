package entities

import (
	"time"
)

type Zone struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Type        string    `json:"type" bson:"type"` // e.g., "storage", "receiving", "shipping"
	Coordinates []float64 `json:"coordinates" bson:"coordinates"` // e.g., [x1, y1, x2, y2] for a rectangular zone
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}
