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

// func (m *MockInventoryService) MoveMaterial(ctx context.Context, cmd services.MoveMaterialCommand) error {
// 	args := m.Called(ctx, cmd)
// 	return args.Error(0)
// }

func TestMoveMaterialCommandHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := commands.NewMoveMaterialCommandHandler(mockService)

	ctx := context.Background()
	cmd := commands.MoveMaterialCommand{
		FromSlotID: "from-slot-id",
		ToSlotID:   "to-slot-id",
		OperatorID: "test-operator-id",
		Reason:     "test-reason",
	}

	// Expect the MoveMaterial method to be called once with the specified arguments
	mockService.On("MoveMaterial", ctx, mock.Anything).Return(nil).Once()

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
