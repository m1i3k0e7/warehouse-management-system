
package entities

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// FailedEvent represents an event that failed to be processed after several retries
// and has been moved to the dead-letter queue (stored in the database).
type FailedEvent struct {
	ID              string         `json:"id" gorm:"primaryKey"`
	Topic           string         `json:"topic"`
	EventType       string         `json:"event_type"`
	Payload         datatypes.JSON `json:"payload"`
	Error           string         `json:"error"`
	CreatedAt       time.Time      `json:"created_at"`
	Resolved        bool           `json:"resolved" gorm:"default:false"`
	ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
	ResolutionNotes string         `json:"resolution_notes,omitempty"`
}

func (FailedEvent) TableName() string {
	return "failed_events"
}

// NewFailedEvent creates a new FailedEvent instance.
func NewFailedEvent(id, topic, eventType string, payload interface{}, err error) (*FailedEvent, error) {
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return &FailedEvent{
		ID:        id,
		Topic:     topic,
		EventType: eventType,
		Payload:   payloadBytes,
		Error:     err.Error(),
		CreatedAt: time.Now(),
	}, nil
}
