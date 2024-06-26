package services

import (
	"bytes"
	"encoding/csv"
	"golang/pkg/models"
	repositoryMocks "golang/pkg/repository/mocks"
	serviceMocks "golang/pkg/services/mocks"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockResponseWriter struct {
	headers map[string][]string
	body    bytes.Buffer
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	return m.body.Write(data)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {}

func TestGetTransactions(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	filters := map[string]interface{}{"status": "completed"}
	transactions := []models.Transaction{{TransactionID: 1, Status: "completed"}}
	mockRepo.On("GetTransactions", filters, 10, 0).Return(transactions, nil)

	result, err := service.GetTransactions(filters, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, transactions, result)
	mockRepo.AssertExpectations(t)
}

func TestGetTransactions_EmptyResults(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	filters := map[string]interface{}{"status": "completed"}
	mockRepo.On("GetTransactions", filters, 10, 0).Return([]models.Transaction{}, nil)

	result, err := service.GetTransactions(filters, 10, 0)
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestCreateTransactionInBatches(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	transactions := []*models.Transaction{{TransactionID: 1}}
	mockRepo.On("CreateTransactionInBatches", transactions, 2).Return(nil)

	err := service.CreateTransactionInBatches(transactions, 2)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateTransactionInBatches_EmptyTransactions(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	transactions := []*models.Transaction{}
	mockRepo.On("CreateTransactionInBatches", transactions, 2).Return(nil)

	err := service.CreateTransactionInBatches(transactions, 2)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProcessCSVFile(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	headers := []string{"TransactionID", "Status"}
	records := [][]string{
		{"1", "completed"},
		{"2", "failed"},
	}
	transactions := []*models.Transaction{
		{TransactionID: 1, Status: "completed"},
		{TransactionID: 2, Status: "failed"},
	}

	reader := csv.NewReader(bytes.NewBufferString("TransactionID,Status\n1,completed\n2,failed\n"))

	csvService.On("ReadCSVHeaders", reader).Return(headers, nil)
	csvService.On("ReadCSVRecord", mock.Anything).Return(records[0], nil).Once()
	csvService.On("ReadCSVRecord", mock.Anything).Return(records[1], nil).Once()
	csvService.On("ReadCSVRecord", mock.AnythingOfType("*csv.Reader")).Return([]string{}, io.EOF)

	mockRepo.On("CreateTransactionInBatches", transactions, 2).Return(nil)

	err := service.ProcessCSVFile(reader)
	assert.NoError(t, err)
	csvService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestProcessCSVFile_EmptyFile(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	reader := csv.NewReader(bytes.NewBufferString(""))

	csvService.On("ReadCSVHeaders", reader).Return([]string{}, io.EOF)

	err := service.ProcessCSVFile(reader)
	assert.Error(t, err)
	assert.Equal(t, "failed to read CSV header: EOF", err.Error())
	csvService.AssertExpectations(t)
}

func TestProcessCSVFile_InvalidData(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	headers := []string{"TransactionID", "Status"}
	reader := csv.NewReader(bytes.NewBufferString("TransactionID,Status\ninvalid_data,completed\n"))

	csvService.On("ReadCSVHeaders", reader).Return(headers, nil)
	csvService.On("ReadCSVRecord", mock.Anything).Return([]string{"invalid_data", "completed"}, nil).Once()
	csvService.On("ReadCSVRecord", mock.AnythingOfType("*csv.Reader")).Return([]string{}, io.EOF)

	err := service.ProcessCSVFile(reader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing record")
	csvService.AssertExpectations(t)
}

func TestExportTransactionsCSV(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	filters := map[string]interface{}{"status": "completed"}
	transactions := []models.Transaction{{TransactionID: 1, Status: "completed"}}
	headers := []string{"TransactionID", "Status"}

	csvService.On("GetCSVHeaders", models.Transaction{}).Return(headers, nil)
	mockRepo.On("GetTransactions", filters, 500, 0).Return(transactions, nil).Once()
	mockRepo.On("GetTransactions", filters, 500, 500).Return([]models.Transaction{}, nil).Once()

	w := &mockResponseWriter{headers: make(map[string][]string)}
	ctx, _ := gin.CreateTestContext(w)

	err := service.ExportTransactionsCSV(ctx, filters)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExportTransactionsCSV_NoTransactions(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	filters := map[string]interface{}{"status": "completed"}
	headers := []string{"TransactionID", "Status"}

	csvService.On("GetCSVHeaders", models.Transaction{}).Return(headers, nil)
	mockRepo.On("GetTransactions", filters, 500, 0).Return([]models.Transaction{}, nil).Once()

	w := &mockResponseWriter{headers: make(map[string][]string)}
	ctx, _ := gin.CreateTestContext(w)

	err := service.ExportTransactionsCSV(ctx, filters)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExportTransactionsJSON(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	filters := map[string]interface{}{"status": "completed"}
	transactions := []models.Transaction{{TransactionID: 1, Status: "completed"}}

	mockRepo.On("GetTransactions", filters, 500, 0).Return(transactions, nil).Once()
	mockRepo.On("GetTransactions", filters, 500, 500).Return([]models.Transaction{}, nil).Once()

	w := &mockResponseWriter{headers: make(map[string][]string)}
	ctx, _ := gin.CreateTestContext(w)

	err := service.ExportTransactionsJSON(ctx, filters)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestExportTransactionsJSON_NoTransactions(t *testing.T) {
	mockRepo := new(repositoryMocks.MockTransactionRepository)
	csvService := new(serviceMocks.MockCSVService)
	service := NewTransactionService(mockRepo, csvService)

	filters := map[string]interface{}{"status": "completed"}

	mockRepo.On("GetTransactions", filters, 500, 0).Return([]models.Transaction{}, nil).Once()

	w := &mockResponseWriter{headers: make(map[string][]string)}
	ctx, _ := gin.CreateTestContext(w)

	err := service.ExportTransactionsJSON(ctx, filters)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
