package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"warehouse/internal/domain/entities"
	"warehouse/internal/infrastructure/database/repositories"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	// Setup test database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_DB_PORT"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&entities.Material{}, &entities.Slot{}, &entities.Operation{}, &entities.Alert{}, &entities.FailedEvent{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// Run tests
	code := m.Run()

	// Teardown
	sqlDB, _ := db.DB()
	sqlDB.Close()

	os.Exit(code)
}

func TestMaterialRepository_Create(t *testing.T) {
	repo := repositories.NewMaterialRepository(db)
	ctx := context.Background()

	material := &entities.Material{
		ID:        "test-material-1",
		Barcode:   "BARCODE-001",
		Name:      "Test Material 1",
		Type:      "CPU",
		Status:    entities.MaterialStatusAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, material)
	assert.NoError(t, err)

	foundMaterial, err := repo.GetByID(ctx, material.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundMaterial)
	assert.Equal(t, material.Barcode, foundMaterial.Barcode)
}

func TestMaterialRepository_GetByBarcode(t *testing.T) {
	repo := repositories.NewMaterialRepository(db)
	ctx := context.Background()

	material := &entities.Material{
		ID:        "test-material-2",
		Barcode:   "BARCODE-002",
		Name:      "Test Material 2",
		Type:      "GPU",
		Status:    entities.MaterialStatusAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, material)
	assert.NoError(t, err)

	foundMaterial, err := repo.GetByBarcode(ctx, material.Barcode)
	assert.NoError(t, err)
	assert.NotNil(t, foundMaterial)
	assert.Equal(t, material.ID, foundMaterial.ID)
}

func TestMaterialRepository_Update(t *testing.T) {
	repo := repositories.NewMaterialRepository(db)
	ctx := context.Background()

	material := &entities.Material{
		ID:        "test-material-3",
		Barcode:   "BARCODE-003",
		Name:      "Test Material 3",
		Type:      "RAM",
		Status:    entities.MaterialStatusAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, material)
	assert.NoError(t, err)

	material.Status = entities.MaterialStatusInUse
	err = repo.Update(ctx, material)
	assert.NoError(t, err)

	foundMaterial, err := repo.GetByID(ctx, material.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundMaterial)
	assert.Equal(t, entities.MaterialStatusInUse, foundMaterial.Status)
}
