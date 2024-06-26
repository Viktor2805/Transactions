package controllers

import (
	"encoding/csv"
	"golang/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	service *services.TransactionService
}

func NewTransactionController(service *services.TransactionService) *TransactionController {
	return &TransactionController{service: service}
}

// UploadCSVFile uploads a CSV file.
// @Summary Upload CSV file
// @Description Uploads a CSV file containing transactions
// @Tags transactions
// @Accept multipart/form-data
// @Produce application/json
// @Param file formData file true "CSV file"
// @Success 200 {object} apierror.SuccessResponse
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /transactions/upload [post]
func (c *TransactionController) UploadCSVFile(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
		return
	}
	defer file.Close()

	if header.Header.Get("Content-Type") != "text/csv" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Please upload a CSV file"})
		return
	}

	reader := csv.NewReader(file)

	err = c.service.ProcessCSVFile(reader)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "File uploaded successfully"})
}

// ExportTransactionsCSV exports transactions as a CSV file.
// @Summary Export transactions as CSV
// @Description Exports transactions based on filters as a CSV file
// @Tags transactions
// @Produce text/csv
// @Param transaction_id query int false "Transaction ID"
// @Param terminal_id query string false "Terminal ID"
// @Param payment_narrative query string false "Payment Narrative"
// @Param date_from query string false "Start Date"
// @Param date_to query string false "End Date"
// @Success 200 {string} string "File transfer"
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /transactions/export-csv [get]
func (c *TransactionController) ExportTransactionsCSV(ctx *gin.Context) {
	filters, err := c.service.ParseFilters(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename=transactions.csv")
	ctx.Header("Content-Type", "text/csv")

	if err := c.service.ExportTransactionsCSV(ctx, filters); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}
}

// ExportTransactionsJSON exports transactions as a JSON file.
// @Summary Export transactions as JSON
// @Description Exports transactions based on filters as a JSON file
// @Tags transactions
// @Produce application/json
// @Param transaction_id query int false "Transaction ID"
// @Param terminal_id query string false "Terminal ID"
// @Param payment_narrative query string false "Payment Narrative"
// @Param date_from query string false "Start Date"
// @Param date_to query string false "End Date"
// @Success 200 {string} string "File transfer"
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /transactions/export-json [get]
func (c *TransactionController) ExportTransactionsJSON(ctx *gin.Context) {
	filters, err := c.service.ParseFilters(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename=transactions.json")
	ctx.Header("Content-Type", "application/json")

	if err := c.service.ExportTransactionsJSON(ctx, filters); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}
}
