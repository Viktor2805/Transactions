package mocks

import (
	"golang/pkg/models"

	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetTransactions(filters map[string]interface{}, limit, offset int) ([]models.Transaction, error) {
	args := m.Called(filters, limit, offset)
	return args.Get(0).([]models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) CreateTransactionInBatches(transactions []*models.Transaction, batchSize int) error {
	args := m.Called(transactions, batchSize)
	return args.Error(0)
}

