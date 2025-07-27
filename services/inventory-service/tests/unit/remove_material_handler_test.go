package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"warehouse/internal/application/commands"
	"warehouse/internal/domain/services"
)

// MockInventoryService is a mock type for the InventoryService
// (Assuming this is defined once in a common test file or similar)
// type MockInventoryService struct {
// 	mock.Mock
// }

// func (m *MockInventoryService) RemoveMaterial(ctx context.Context, cmd services.RemoveMaterialCommand) error {
// 	args := m.Called(ctx, cmd)
// 	return args.Error(0)
// }

func TestRemoveMaterialCommandHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := commands.NewRemoveMaterialCommandHandler(mockService)

	ctx := context.Background()
	cmd := commands.RemoveMaterialCommand{
		SlotID:     "test-slot-id",
		OperatorID: "test-operator-id",
		Reason:     "test-reason",
	}

	// Expect the RemoveMaterial method to be called once with the specified arguments
	mockService.On("RemoveMaterial", ctx, mock.Anything).Return(nil).Once()

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
