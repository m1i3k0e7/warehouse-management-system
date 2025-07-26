package services

import (
    "context"
    "fmt"
    
    "warehouse/internal/domain/entities"
    "warehouse/pkg/logger"
)

type AlertService struct {
    eventService *EventService
}

func NewAlertService(eventService *EventService) *AlertService {
    return &AlertService{eventService: eventService}
}

func (s *AlertService) SendShelfHealthAlert(ctx context.Context, health *entities.ShelfHealth) {
    alert := map[string]interface{}{
        "type":         "shelf_health",
        "shelf_id":     health.ShelfID,
        "health_score": health.HealthScore,
        "message":      fmt.Sprintf("Shelf %s health score is %.2f%%", health.ShelfID, health.HealthScore),
        "severity":     s.determineSeverity(health.HealthScore),
        "timestamp":    health.LastCheckTime,
    }
    
    if err := s.eventService.PublishEvent(ctx, "alert.shelf_health", alert); err != nil {
        logger.Error("Failed to send shelf health alert", err)
    }
}

func (s *AlertService) determineSeverity(healthScore float64) string {
    if healthScore < 80 {
        return "critical"
    } else if healthScore < 90 {
        return "high"
    } else if healthScore < 95 {
        return "medium"
    }
    return "low"
}