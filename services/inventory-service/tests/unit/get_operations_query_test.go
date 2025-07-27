package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"warehouse/internal/application/queries"
	"warehouse/internal/domain/entities"
	"warehouse/internal/domain/repositories"
)

// MockOperationRepository is a mock type for the OperationRepository
type MockOperationRepository struct {
	mock.Mock
}

// List is a mock method
func (m *MockOperationRepository) List(ctx context.Context, limit, offset int) ([]*entities.Operation, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Operation), args.Error(1)
}

// Implement other methods of OperationRepository if needed by other tests
func (m *MockOperationRepository) Create(ctx context.Context, operation *entities.Operation) error {
	args := m.Called(ctx, operation)
	return args.Error(0)
}
func (m *MockOperationRepository) CreateWithTx(ctx context.Context, tx interface{}, operation *entities.Operation) error {
	args := m.Called(ctx, tx, operation)
	return args.Error(0)
}
func (m *MockOperationRepository) GetByID(ctx context.Context, id string) (*entities.Operation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Operation), args.Error(1)
}
func (m *MockOperationRepository) GetByShelfID(ctx context.Context, shelfID string, limit, offset int) ([]*entities.Operation, error) {
	args := m.Called(ctx, shelfID, limit, offset)
	return args.Get(0).([]*entities.Operation), args.Error(1)
}
func (m *MockOperationRepository) GetByOperatorID(ctx context.Context, operatorID string, limit, offset int) ([]*entities.Operation, error) {
	args := m.Called(ctx, operatorID, limit, offset)
	return args.Get(0).([]*entities.Operation), args.Error(1)
}

func TestGetOperationsQueryHandler_Handle(t *testing.T) {
	// Arrange
	mockRepo := new(MockOperationRepository)
	handler := queries.NewGetOperationsQueryHandler(mockRepo)

	ctx := context.Background()
	query := queries.GetOperationsQuery{
		Limit:  10,
		Offset: 0,
	}

	expectedOperations := []*entities.Operation{
		{ID: "op1", Type: "placement"},
		{ID: "op2", Type: "removal"},
	}

	// Expect the List method to be called once with the specified arguments
	mockRepo.On("List", ctx, query.Limit, query.Offset).Return(expectedOperations, nil).Once()

	// Act
	operations, err := handler.Handle(ctx, query)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedOperations, operations)
	mockRepo.AssertExpectations(t)
}
