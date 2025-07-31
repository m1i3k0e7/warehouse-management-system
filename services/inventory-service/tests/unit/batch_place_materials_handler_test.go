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

// func (m *MockInventoryService) BatchPlaceMaterials(ctx context.Context, cmds []services.PlaceMaterialCommand) error {
// 	args := m.Called(ctx, cmds)
// 	return args.Error(0)
// }

func TestBatchPlaceMaterialsCommandHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := commands.NewBatchPlaceMaterialsCommandHandler(mockService)

	ctx := context.Background()
	cmd := commands.BatchPlaceMaterialsCommand{
		Commands: []services.PlaceMaterialCommand{
			{MaterialBarcode: "mat1", SlotID: "slot1", OperatorID: "op1"},
			{MaterialBarcode: "mat2", SlotID: "slot2", OperatorID: "op1"},
		},
	}

	// Expect the BatchPlaceMaterials method to be called once with the specified arguments
	mockService.On("BatchPlaceMaterials", ctx, mock.Anything).Return(nil).Once()

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
