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

// func (m *MockInventoryService) ReserveSlots(ctx context.Context, cmd services.ReserveSlotsCommand) error {
// 	args := m.Called(ctx, cmd)
// 	return args.Error(0)
// }

func TestReserveSlotsCommandHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := commands.NewReserveSlotsCommandHandler(mockService)

	ctx := context.Background()
	cmd := commands.ReserveSlotsCommand{
		SlotIDs:    []string{"slot1", "slot2"},
		OperatorID: "test-operator-id",
		Duration:   60,
		Purpose:    "maintenance",
	}

	// Expect the ReserveSlots method to be called once with the specified arguments
	mockService.On("ReserveSlots", ctx, mock.Anything).Return(nil).Once()

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
