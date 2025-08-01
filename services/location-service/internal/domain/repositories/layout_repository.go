package repositories

import (
	"context"
	"github.com/your-repo/wms/location-service/internal/domain/entities"
)

// LayoutRepository defines the interface for interacting with the overall warehouse layout.
type LayoutRepository interface {
	FindZoneByID(ctx context.Context, id string) (*entities.Zone, error)
	SaveZone(ctx context.Context, zone *entities.Zone) error
	FindAllShelvesInZone(ctx context.Context, zoneID string) ([]*entities.Shelf, error)
}