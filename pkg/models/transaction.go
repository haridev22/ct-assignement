package models

import (
	"time"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TypeEthTransfer     TransactionType = "ETH_TRANSFER"
	TypeERC20Transfer   TransactionType = "ERC20_TRANSFER"
	TypeERC721Transfer  TransactionType = "ERC721_TRANSFER"
	TypeERC1155Transfer TransactionType = "ERC1155_TRANSFER"
	TypeContractCall    TransactionType = "CONTRACT_CALL"
	TypeInternalTx      TransactionType = "INTERNAL_TRANSFER"
)

// Transaction represents a processed transaction ready for CSV export
type Transaction struct {
	Hash              string        `json:"hash"`
	Timestamp         time.Time     `json:"timestamp"`
	From              string        `json:"from"`
	To                string        `json:"to"`
	Type              TransactionType `json:"type"`
	AssetContractAddr string        `json:"asset_contract_address,omitempty"`
	AssetSymbol       string        `json:"asset_symbol,omitempty"`
	TokenID           string        `json:"token_id,omitempty"`
	Value             string        `json:"value"`
	GasFee            string        `json:"gas_fee"`
}

// CSVRecord converts a transaction to a slice of strings for CSV output
func (t *Transaction) CSVRecord() []string {
	return []string{
		t.Hash,
		t.Timestamp.Format(time.RFC3339),
		t.From,
		t.To,
		string(t.Type),
		t.AssetContractAddr,
		t.AssetSymbol,
		t.TokenID,
		t.Value,
		t.GasFee,
	}
}

// CSVHeaders returns the CSV header row
func CSVHeaders() []string {
	return []string{
		"Transaction Hash",
		"Date & Time",
		"From Address",
		"To Address",
		"Transaction Type",
		"Asset Contract Address",
		"Asset Symbol / Name",
		"Token ID",
		"Value / Amount",
		"Gas Fee (ETH)",
	}
}
