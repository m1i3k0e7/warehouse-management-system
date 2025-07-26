package repositories

import (
    "context"
    "inventory-service/internal/domain/entities"
)

type AlertRepository interface {
    Create(ctx context.Context, alert *entities.Alert) error
    GetByID(ctx context.Context, id string) (*entities.Alert, error)
    GetActiveAlerts(ctx context.Context, limit, offset int) ([]*entities.Alert, error)
    GetByShelfID(ctx context.Context, shelfID string, limit, offset int) ([]*entities.Alert, error)
    UpdateStatus(ctx context.Context, id string, status string) error
    MarkAsResolved(ctx context.Context, id string) error
    List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.Alert, error)
}