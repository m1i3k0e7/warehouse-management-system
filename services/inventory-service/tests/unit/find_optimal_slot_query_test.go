package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"WMS/services/inventory-service/internal/application/queries"
	"WMS/services/inventory-service/internal/domain/entities"
	"WMS/services/inventory-service/internal/domain/services"
)

// MockInventoryService is a mock type for the InventoryService
// (Assuming this is defined once in a common test file or similar)
// type MockInventoryService struct {
// 	mock.Mock
// }

// func (m *MockInventoryService) FindOptimalSlot(ctx context.Context, materialType, shelfID string) (*entities.Slot, error) {
// 	args := m.Called(ctx, materialType, shelfID)
// 	return args.Get(0).(*entities.Slot), args.Error(1)
// }

func TestFindOptimalSlotQueryHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := queries.NewFindOptimalSlotQueryHandler(mockService)

	ctx := context.Background()
	query := queries.FindOptimalSlotQuery{
		MaterialType: "CPU",
		ShelfID:      "shelf-1",
	}

	expectedSlot := &entities.Slot{
		ID:      "slot-1",
		ShelfID: "shelf-1",
		Row:     1,
		Column:  1,
		Status:  entities.SlotStatusEmpty,
	}

	// Expect the FindOptimalSlot method to be called once with the specified arguments
	mockService.On("FindOptimalSlot", ctx, query.MaterialType, query.ShelfID).Return(expectedSlot, nil).Once()

	// Act
	slot, err := handler.Handle(ctx, query)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedSlot, slot)
	mockService.AssertExpectations(t)
}
