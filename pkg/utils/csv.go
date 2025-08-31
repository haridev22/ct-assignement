package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"eth-tx-history/pkg/models"
)

// ExportTransactionsToCSV writes transactions to a CSV file
func ExportTransactionsToCSV(transactions []models.Transaction, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	if err := writer.Write(models.CSVHeaders()); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write transaction records
	for _, tx := range transactions {
		if err := writer.Write(tx.CSVRecord()); err != nil {
			return fmt.Errorf("failed to write transaction record: %w", err)
		}
	}

	return nil
}
