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

// func (m *MockInventoryService) SearchMaterials(ctx context.Context, query string, limit, offset int) ([]*entities.Material, error) {
// 	args := m.Called(ctx, query, limit, offset)
// 	return args.Get(0).([]*entities.Material), args.Error(1)
// }

func TestSearchMaterialsQueryHandler_Handle(t *testing.T) {
	// Arrange
	mockService := new(MockInventoryService)
	handler := queries.NewSearchMaterialsQueryHandler(mockService)

	ctx := context.Background()
	query := queries.SearchMaterialsQuery{
		Query:  "test",
		Limit:  10,
		Offset: 0,
	}

	expectedMaterials := []*entities.Material{
		{ID: "mat1", Name: "Test Material 1"},
		{ID: "mat2", Name: "Test Material 2"},
	}

	// Expect the SearchMaterials method to be called once with the specified arguments
	mockService.On("SearchMaterials", ctx, query.Query, query.Limit, query.Offset).Return(expectedMaterials, nil).Once()

	// Act
	materials, err := handler.Handle(ctx, query)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedMaterials, materials)
	mockService.AssertExpectations(t)
}
