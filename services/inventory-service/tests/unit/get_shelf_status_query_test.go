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

// func (m *MockInventoryService) GetShelfStatus(ctx context.Context, shelfID string) (*entities.ShelfStatus, error) {
// 	args := m.Called(ctx, shelfID)
// 	return args.Get(0).(*entities.ShelfStatus), args.Error(1)
// }

func TestGetShelfStatusQueryHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := queries.NewGetShelfStatusQueryHandler(mockService)

	ctx := context.Background()
	query := queries.GetShelfStatusQuery{
		ShelfID: "test-shelf-id",
	}

	expectedStatus := &entities.ShelfStatus{
		ShelfID:       "test-shelf-id",
		TotalSlots:    100,
		EmptySlots:    50,
		OccupiedSlots: 50,
	}

	// Expect the GetShelfStatus method to be called once with the specified arguments
	mockService.On("GetShelfStatus", ctx, query.ShelfID).Return(expectedStatus, nil).Once()

	// Act
	status, err := handler.Handle(ctx, query)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
	mockService.AssertExpectations(t)
}
