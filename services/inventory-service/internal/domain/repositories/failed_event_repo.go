
package repositories

import (
	"context"
	"inventory-service/internal/domain/entities"
)

// FailedEventRepository defines the interface for interacting with the failed_events table.
type FailedEventRepository interface {
	Create(ctx context.Context, event *entities.FailedEvent) error
	GetByID(ctx context.Context, id string) (*entities.FailedEvent, error)
	ListUnresolved(ctx context.Context, limit, offset int) ([]*entities.FailedEvent, error)
	MarkAsResolved(ctx context.Context, id, notes string) error
}
