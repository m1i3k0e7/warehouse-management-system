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

// func (m *MockInventoryService) HandleSlotError(ctx context.Context, slotID, errorType string) error {
// 	args := m.Called(ctx, slotID, errorType)
// 	return args.Error(0)
// }

func TestHandleSlotErrorCommandHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := commands.NewHandleSlotErrorCommandHandler(mockService)

	ctx := context.Background()
	cmd := commands.HandleSlotErrorCommand{
		SlotID:    "test-slot-id",
		ErrorType: "sensor_error",
	}

	// Expect the HandleSlotError method to be called once with the specified arguments
	mockService.On("HandleSlotError", ctx, cmd.SlotID, cmd.ErrorType).Return(nil).Once()

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
