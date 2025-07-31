package entities

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type AlertType string

const (
	AlertTypeShelfHealth AlertType = "shelf_health"
	AlertTypeSlotError   AlertType = "slot_error"
	AlertTypeSystem      AlertType = "system_alert"
)

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

type Alert struct {
	ID         string      `json:"id" gorm:"primaryKey"`
	Type       AlertType   `json:"type"`
	ShelfID    string      `json:"shelf_id,omitempty"`
	SlotID     string      `json:"slot_id,omitempty"`
	Message    string      `json:"message"`
	Severity   AlertSeverity `json:"severity"`
	Status     AlertStatus   `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	ResolvedAt *time.Time  `json:"resolved_at,omitempty"`
	Metadata   JSON        `json:"metadata" gorm:"type:jsonb"`
}

type SystemAlertEvent struct {
	ID    		string    `json:"alert_id"`
	Type  		AlertType `json:"alert_type"`
	Severity    AlertSeverity `json:"severity"`
	Message 	string    `json:"message"`
	Timestamp 	time.Time `json:"timestamp"`
	Details  	JSON      `json:"details" gorm:"type:jsonb"`
}
	

// JSON type for GORM
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

func (Alert) TableName() string {
	return "alerts"
}