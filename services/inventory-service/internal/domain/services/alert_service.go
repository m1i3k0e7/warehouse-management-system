/* 
    * AlertService is responsible for sending alerts based on shelf health checks.
    * It uses the EventService to publish alerts to a message broker.
    * It determines the severity of the alert based on the health score of the shelf.
*/
package services

import (
	"context"
	"fmt"

	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/pkg/utils/logger"
)

type AlertService struct {
	eventService *EventService
}

func NewAlertService(eventService *EventService) *AlertService {
	return &AlertService{eventService: eventService}
}

func (s *AlertService) SendShelfHealthAlert(ctx context.Context, health *entities.ShelfHealth) {
	alert := map[string]interface{}{
		"type":         entities.AlertTypeShelfHealth,
		"shelf_id":     health.ShelfID,
		"health_score": health.HealthScore,
		"message":      fmt.Sprintf("Shelf %s health score is %.2f%%", health.ShelfID, health.HealthScore),
		"severity":     s.determineSeverity(health.HealthScore),
		"timestamp":    health.LastCheckTime,
	}

	if err := s.eventService.PublishEvent(ctx, EventTypeShelfHealthAlert, alert); err != nil {
		logger.Error("Failed to send shelf health alert", err)
	}
}

func (s *AlertService) determineSeverity(healthScore float64) entities.AlertSeverity {
	if healthScore < 80 {
		return entities.AlertSeverityCritical
	} else if healthScore < 90 {
		return entities.AlertSeverityHigh
	} else if healthScore < 95 {
		return entities.AlertSeverityMedium
	}
	return entities.AlertSeverityLow
}