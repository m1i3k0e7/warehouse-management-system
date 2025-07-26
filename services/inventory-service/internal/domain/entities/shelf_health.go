package entities

import "time"

type ShelfHealth struct {
    ShelfID           string    `json:"shelf_id"`
    TotalSlots        int       `json:"total_slots"`
    HealthySlots      int       `json:"healthy_slots"`
    ErrorSlots        int       `json:"error_slots"`
    MaintenanceSlots  int       `json:"maintenance_slots"`
    HealthScore       float64   `json:"health_score"`
    LastCheckTime     time.Time `json:"last_check_time"`
    Issues            []string  `json:"issues,omitempty"`
}