package repositories

import (
	"context"

	"warehouse/location-service/internal/domain/entities"
)

type LayoutRepository interface {
	CreateZone(ctx context.Context, zone *entities.Zone) error
	GetZoneByID(ctx context.Context, id string) (*entities.Zone, error)
	GetAllZones(ctx context.Context) ([]entities.Zone, error)
	UpdateZone(ctx context.Context, zone *entities.Zone) error
	DeleteZone(ctx context.Context, id string) error

	CreatePath(ctx context.Context, path *entities.Path) error
	GetPathByID(ctx context.Context, id string) (*entities.Path, error)
	GetPathsBySlots(ctx context.Context, startSlotID, endSlotID string) ([]entities.Path, error)
	UpdatePath(ctx context.Context, path *entities.Path) error
	DeletePath(ctx context.Context, id string) error
}
