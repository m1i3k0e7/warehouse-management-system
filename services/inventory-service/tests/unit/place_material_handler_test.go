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
type MockInventoryService struct {
	mock.Mock
}

// PlaceMaterial is a mock method
func (m *MockInventoryService) PlaceMaterial(ctx context.Context, cmd services.PlaceMaterialCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func TestPlaceMaterialCommandHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := commands.NewPlaceMaterialCommandHandler(mockService)

	ctx := context.Background()
	cmd := commands.PlaceMaterialCommand{
		MaterialBarcode: "test-barcode",
		SlotID:          "test-slot-id",
		OperatorID:      "test-operator-id",
	}

	// Expect the PlaceMaterial method to be called once with the specified arguments
	mockService.On("PlaceMaterial", ctx, mock.Anything).Return(nil).Once()

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
