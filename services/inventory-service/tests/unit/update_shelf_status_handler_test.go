package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"WMS/services/inventory-service/internal/application/commands"
	"WMS/services/inventory-service/internal/domain/services"
)

// MockInventoryService is a mock type for the InventoryService
// (Assuming this is defined once in a common test file or similar)
// type MockInventoryService struct {
// 	mock.Mock
// }

// func (m *MockInventoryService) UpdateShelfStatus(ctx context.Context, shelfID, status string) error {
// 	args := m.Called(ctx, shelfID, status)
// 	return args.Error(0)
// }

func TestUpdateShelfStatusCommandHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := commands.NewUpdateShelfStatusCommandHandler(mockService)

	ctx := context.Background()
	cmd := commands.UpdateShelfStatusCommand{
		ShelfID: "test-shelf-id",
		Status:  "online",
	}

	// Expect the UpdateShelfStatus method to be called once with the specified arguments
	mockService.On("UpdateShelfStatus", ctx, cmd.ShelfID, cmd.Status).Return(nil).Once()

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
