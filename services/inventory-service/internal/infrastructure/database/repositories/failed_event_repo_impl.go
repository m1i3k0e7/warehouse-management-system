
package repositories

import (
	"context"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/domain/repositories"
	"time"

	"gorm.io/gorm"
)

type failedEventRepository struct {
	db *gorm.DB
}

// NewFailedEventRepository creates a new instance of FailedEventRepository.
func NewFailedEventRepository(db *gorm.DB) repositories.FailedEventRepository {
	return &failedEventRepository{db: db}
}

func (r *failedEventRepository) Create(ctx context.Context, event *entities.FailedEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *failedEventRepository) GetByID(ctx context.Context, id string) (*entities.FailedEvent, error) {
	var event entities.FailedEvent
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	return &event, err
}

func (r *failedEventRepository) ListUnresolved(ctx context.Context, limit, offset int) ([]*entities.FailedEvent, error) {
	var events []*entities.FailedEvent
	err := r.db.WithContext(ctx).
		Where("resolved = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

func (r *failedEventRepository) MarkAsResolved(ctx context.Context, id, notes string) error {
	return r.db.WithContext(ctx).Model(&entities.FailedEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"resolved":         true,
		"resolved_at":      time.Now(),
		"resolution_notes": notes,
	}).Error
}
