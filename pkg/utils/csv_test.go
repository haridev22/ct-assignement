package utils

import (
	"encoding/csv"
	"os"
	"testing"
	"time"

	"eth-tx-history/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestExportTransactionsToCSV(t *testing.T) {
	// Create temporary directory for test output
	tempDir, err := os.MkdirTemp("", "csv-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test transactions
	transactions := []models.Transaction{
		{
			Hash:              "0x123abc",
			Timestamp:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			From:              "0xsender1",
			To:                "0xreceiver1",
			Type:              models.TypeEthTransfer,
			Value:             "1.500000000000000000",
			GasFee:            "0.000210000000000000",
		},
		{
			Hash:              "0x456def",
			Timestamp:         time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			From:              "0xsender2",
			To:                "0xreceiver2",
			Type:              models.TypeERC20Transfer,
			AssetContractAddr: "0xtoken",
			AssetSymbol:       "USDC",
			Value:             "100.000000",
			GasFee:            "0.000650000000000000",
		},
		{
			Hash:              "0x789ghi",
			Timestamp:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
			From:              "0xsender3",
			To:                "0xreceiver3",
			Type:              models.TypeERC721Transfer,
			AssetContractAddr: "0xnft",
			AssetSymbol:       "BAYC",
			TokenID:           "1234",
			Value:             "1",
			GasFee:            "0.001200000000000000",
		},
	}

	// Generate file path
	outputPath := tempDir + "/transactions_export.csv"
	
	// Export transactions
	err = ExportTransactionsToCSV(transactions, outputPath)
	assert.NoError(t, err)
	assert.FileExists(t, outputPath)

	// Verify CSV content
	file, err := os.Open(outputPath)
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Check header
	assert.Equal(t, models.CSVHeaders(), records[0])
	
	// Check number of rows (header + 3 transactions)
	assert.Len(t, records, 4)
	
	// Check specific record values
	assert.Equal(t, "0x123abc", records[1][0]) // Hash of first transaction
	assert.Equal(t, "0xsender1", records[1][2]) // From address of first transaction
	assert.Equal(t, "USDC", records[2][6]) // Token symbol of second transaction
	assert.Equal(t, "1234", records[3][7]) // Token ID of third transaction
}

func TestExportTransactionsToCSV_EmptyList(t *testing.T) {
	// Create temporary directory for test output
	tempDir, err := os.MkdirTemp("", "csv-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Generate file path
	outputPath := tempDir + "/empty_transactions.csv"
	
	// Test with empty transaction list
	err = ExportTransactionsToCSV([]models.Transaction{}, outputPath)
	assert.NoError(t, err)
	assert.FileExists(t, outputPath)

	// Verify CSV has only header row
	file, err := os.Open(outputPath)
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 1) // Only header row
	assert.Equal(t, models.CSVHeaders(), records[0])
}

func TestExportTransactionsToCSV_InvalidPath(t *testing.T) {
	// Test with invalid output path that requires non-existent directories
	err := ExportTransactionsToCSV([]models.Transaction{}, "/dev/null/impossible/path.csv")
	assert.Error(t, err)
}
