package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"warehouse/internal/application/queries"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/services"
)

// MockInventoryService is a mock type for the InventoryService
// (Assuming this is defined once in a common test file or similar)
// type MockInventoryService struct {
// 	mock.Mock
// }

// func (m *MockInventoryService) HealthCheckShelf(ctx context.Context, shelfID string) (*entities.ShelfHealth, error) {
// 	args := m.Called(ctx, shelfID)
// 	return args.Get(0).(*entities.ShelfHealth), args.Error(1)
// }

func TestHealthCheckShelfQueryHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := queries.NewHealthCheckShelfQueryHandler(mockService)

	ctx := context.Background()
	query := queries.HealthCheckShelfQuery{
		ShelfID: "test-shelf-id",
	}

	expectedHealth := &entities.ShelfHealth{
		ShelfID:      "test-shelf-id",
		HealthScore:  99.5,
		TotalSlots:   700,
		HealthySlots: 699,
	}

	// Expect the HealthCheckShelf method to be called once with the specified arguments
	mockService.On("HealthCheckShelf", ctx, query.ShelfID).Return(expectedHealth, nil).Once()

	// Act
	health, err := handler.Handle(ctx, query)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedHealth, health)
	mockService.AssertExpectations(t)
}
