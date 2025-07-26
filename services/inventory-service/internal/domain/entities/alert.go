package entities

import (
    "time"
    "database/sql/driver"
    "encoding/json"
    "errors"
)

type Alert struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Type      string    `json:"type"`
    ShelfID   string    `json:"shelf_id,omitempty"`
    SlotID    string    `json:"slot_id,omitempty"`
    Message   string    `json:"message"`
    Severity  string    `json:"severity"` // low, medium, high, critical
    Status    string    `json:"status"`   // active, acknowledged, resolved
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    ResolvedAt *time.Time `json:"resolved_at,omitempty"`
    Metadata  JSON      `json:"metadata" gorm:"type:jsonb"`
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