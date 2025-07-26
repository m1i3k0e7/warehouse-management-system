package repositories

import (
	"context"
	"inventory-service/internal/domain/entities"
	"inventory-service/internal/domain/repositories"
	
	"gorm.io/gorm"
)

type materialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) repositories.MaterialRepository {
	return &materialRepository{db: db}
}

func (r *materialRepository) Create(ctx context.Context, material *entities.Material) error {
	return r.db.WithContext(ctx).Create(material).Error
}

func (r *materialRepository) GetByID(ctx context.Context, id string) (*entities.Material, error) {
	var material entities.Material
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&material).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

func (r *materialRepository) GetByBarcode(ctx context.Context, barcode string) (*entities.Material, error) {
	var material entities.Material
	err := r.db.WithContext(ctx).Where("barcode = ?", barcode).First(&material).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

func (r *materialRepository) Update(ctx context.Context, material *entities.Material) error {
	return r.db.WithContext(ctx).Save(material).Error
}

func (r *materialRepository) UpdateWithTx(ctx context.Context, tx *gorm.DB, material *entities.Material) error {
	return tx.WithContext(ctx).Save(material).Error
}

func (r *materialRepository) List(ctx context.Context, limit, offset int) ([]*entities.Material, error) {
	var materials []*entities.Material
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&materials).Error
	return materials, err
}

func (r *materialRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Material, error) {
	var materials []*entities.Material
	err := r.db.WithContext(ctx).
		Where("name ILIKE ? OR barcode ILIKE ? OR type ILIKE ?", 
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&materials).Error
	return materials, err
}