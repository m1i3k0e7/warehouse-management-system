package entities

import (
	"time"
)

type Path struct {
	ID        string    `json:"id" bson:"_id"`
	StartSlot string    `json:"start_slot" bson:"start_slot"`
	EndSlot   string    `json:"end_slot" bson:"end_slot"`
	PathNodes []string  `json:"path_nodes" bson:"path_nodes"` // Ordered list of slot IDs or coordinates
	Distance  float64   `json:"distance" bson:"distance"`
	Duration  time.Duration `json:"duration" bson:"duration"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
