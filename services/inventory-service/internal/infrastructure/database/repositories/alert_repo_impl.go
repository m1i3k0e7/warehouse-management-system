package repositories

import (
    "context"
    "inventory-service/internal/domain/entities"
    "inventory-service/internal/domain/repositories"
    "time"
    
    "gorm.io/gorm"
)

type alertRepository struct {
    db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) repositories.AlertRepository {
    return &alertRepository{db: db}
}

func (r *alertRepository) Create(ctx context.Context, alert *entities.Alert) error {
    return r.db.WithContext(ctx).Create(alert).Error
}

func (r *alertRepository) GetByID(ctx context.Context, id string) (*entities.Alert, error) {
    var alert entities.Alert
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&alert).Error
    if err != nil {
        return nil, err
    }
    return &alert, nil
}

func (r *alertRepository) GetActiveAlerts(ctx context.Context, limit, offset int) ([]*entities.Alert, error) {
    var alerts []*entities.Alert
    err := r.db.WithContext(ctx).
        Where("status IN ?", []string{"active", "acknowledged"}).
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&alerts).Error
    return alerts, err
}

func (r *alertRepository) GetByShelfID(ctx context.Context, shelfID string, limit, offset int) ([]*entities.Alert, error) {
    var alerts []*entities.Alert
    err := r.db.WithContext(ctx).
        Where("shelf_id = ?", shelfID).
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&alerts).Error
    return alerts, err
}

func (r *alertRepository) UpdateStatus(ctx context.Context, id string, status string) error {
    return r.db.WithContext(ctx).
        Model(&entities.Alert{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "status": status,
            "updated_at": time.Now(),
        }).Error
}

func (r *alertRepository) MarkAsResolved(ctx context.Context, id string) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&entities.Alert{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "status": "resolved",
            "resolved_at": &now,
            "updated_at": now,
        }).Error
}

func (r *alertRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.Alert, error) {
    query := r.db.WithContext(ctx)
    
    for key, value := range filters {
        switch key {
        case "severity":
            query = query.Where("severity = ?", value)
        case "status":
            query = query.Where("status = ?", value)
        case "type":
            query = query.Where("type = ?", value)
        case "shelf_id":
            query = query.Where("shelf_id = ?", value)
        case "date_from":
            query = query.Where("created_at >= ?", value)
        case "date_to":
            query = query.Where("created_at <= ?", value)
        }
    }
    
    var alerts []*entities.Alert
    err := query.Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&alerts).Error
    
    return alerts, err
}