package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_CSVRecord(t *testing.T) {
	// Test case: Complete transaction with all fields
	tx := Transaction{
		Hash:              "0xabc123",
		Timestamp:         time.Date(2023, 3, 15, 12, 30, 45, 0, time.UTC),
		From:              "0xsender",
		To:                "0xreceiver",
		Type:              TypeEthTransfer,
		AssetContractAddr: "0xcontract",
		AssetSymbol:       "ETH",
		TokenID:           "42",
		Value:             "1.500000000000000000",
		GasFee:            "0.000210000000000000",
	}

	record := tx.CSVRecord()

	// Check each field in the CSV record
	assert.Equal(t, "0xabc123", record[0], "Transaction hash should match")
	assert.Equal(t, "2023-03-15T12:30:45Z", record[1], "Timestamp format should be RFC3339")
	assert.Equal(t, "0xsender", record[2], "From address should match")
	assert.Equal(t, "0xreceiver", record[3], "To address should match")
	assert.Equal(t, "ETH_TRANSFER", record[4], "Transaction type should match")
	assert.Equal(t, "0xcontract", record[5], "Asset contract address should match")
	assert.Equal(t, "ETH", record[6], "Asset symbol should match")
	assert.Equal(t, "42", record[7], "Token ID should match")
	assert.Equal(t, "1.500000000000000000", record[8], "Value should match")
	assert.Equal(t, "0.000210000000000000", record[9], "Gas fee should match")

	// Test case: Minimal transaction with only required fields
	minimalTx := Transaction{
		Hash:      "0xdef456",
		Timestamp: time.Date(2023, 3, 16, 0, 0, 0, 0, time.UTC),
		From:      "0xminimal",
		To:        "0xminimal",
		Type:      TypeInternalTx,
		Value:     "0.1",
		GasFee:    "0",
	}

	minimalRecord := minimalTx.CSVRecord()
	
	assert.Equal(t, "0xdef456", minimalRecord[0], "Transaction hash should match")
	assert.Equal(t, "2023-03-16T00:00:00Z", minimalRecord[1], "Timestamp format should be RFC3339")
	assert.Equal(t, "0xminimal", minimalRecord[2], "From address should match")
	assert.Equal(t, "0xminimal", minimalRecord[3], "To address should match")
	assert.Equal(t, "INTERNAL_TRANSFER", minimalRecord[4], "Transaction type should match")
	assert.Equal(t, "", minimalRecord[5], "Asset contract address should be empty")
	assert.Equal(t, "", minimalRecord[6], "Asset symbol should be empty")
	assert.Equal(t, "", minimalRecord[7], "Token ID should be empty")
	assert.Equal(t, "0.1", minimalRecord[8], "Value should match")
	assert.Equal(t, "0", minimalRecord[9], "Gas fee should match")
}

func TestCSVHeaders(t *testing.T) {
	headers := CSVHeaders()
	
	// Check the number of headers
	assert.Len(t, headers, 10, "There should be 10 headers")
	
	// Check specific headers
	assert.Equal(t, "Transaction Hash", headers[0])
	assert.Equal(t, "Date & Time", headers[1])
	assert.Equal(t, "From Address", headers[2])
	assert.Equal(t, "To Address", headers[3])
	assert.Equal(t, "Transaction Type", headers[4])
	assert.Equal(t, "Asset Contract Address", headers[5])
	assert.Equal(t, "Asset Symbol / Name", headers[6])
	assert.Equal(t, "Token ID", headers[7])
	assert.Equal(t, "Value / Amount", headers[8])
	assert.Equal(t, "Gas Fee (ETH)", headers[9])
}
