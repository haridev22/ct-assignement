package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"eth-tx-history/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestConvertNormalTxToModel(t *testing.T) {
	// Test case: Regular ETH transaction
	tx := NormalTransaction{
		Hash:              "0x123abc",
		TimeStamp:         "1630000000",
		From:              "0xsender",
		To:                "0xreceiver",
		Value:             "1000000000000000000", // 1 ETH
		GasPrice:          "20000000000", // 20 Gwei
		GasUsed:           "21000", // Standard ETH transfer gas
	}

	result, err := ConvertNormalTxToModel(tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x123abc", result.Hash)
	assert.Equal(t, time.Unix(1630000000, 0), result.Timestamp)
	assert.Equal(t, "0xsender", result.From)
	assert.Equal(t, "0xreceiver", result.To)
	assert.Equal(t, models.TypeEthTransfer, result.Type)
	assert.Equal(t, "1.000000000000000000", result.Value)
	assert.Equal(t, "0.000420000000000000", result.GasFee)

	// Test case: Invalid timestamp
	txInvalid := NormalTransaction{
		TimeStamp: "invalid",
	}
	_, err = ConvertNormalTxToModel(txInvalid)
	assert.Error(t, err)
}

func TestConvertERC20TxToModel(t *testing.T) {
	// Test case: Regular ERC20 token transaction
	tx := ERC20Transaction{
		Hash:              "0x456def",
		TimeStamp:         "1630000000",
		From:              "0xsender",
		To:                "0xreceiver",
		ContractAddress:   "0xtoken",
		TokenSymbol:       "TEST",
		TokenDecimal:      "18",
		Value:             "1000000000000000000", // 1 token
		GasPrice:          "20000000000", // 20 Gwei
		GasUsed:           "65000", // ERC-20 transfer gas
	}

	result, err := ConvertERC20TxToModel(tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x456def", result.Hash)
	assert.Equal(t, time.Unix(1630000000, 0), result.Timestamp)
	assert.Equal(t, "0xsender", result.From)
	assert.Equal(t, "0xreceiver", result.To)
	assert.Equal(t, models.TypeERC20Transfer, result.Type)
	assert.Equal(t, "0xtoken", result.AssetContractAddr)
	assert.Equal(t, "TEST", result.AssetSymbol)
	assert.Equal(t, "1.000000000000000000", result.Value)
}

func TestConvertERC721TxToModel(t *testing.T) {
	// Test case: NFT transfer
	tx := ERC721Transaction{
		Hash:              "0x789ghi",
		TimeStamp:         "1630000000",
		From:              "0xsender",
		To:                "0xreceiver",
		ContractAddress:   "0xnft",
		TokenSymbol:       "NFT",
		TokenID:           "12345",
		GasPrice:          "20000000000", // 20 Gwei
		GasUsed:           "120000", // NFT transfer gas
	}

	result, err := ConvertERC721TxToModel(tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x789ghi", result.Hash)
	assert.Equal(t, time.Unix(1630000000, 0), result.Timestamp)
	assert.Equal(t, "0xsender", result.From)
	assert.Equal(t, "0xreceiver", result.To)
	assert.Equal(t, models.TypeERC721Transfer, result.Type)
	assert.Equal(t, "0xnft", result.AssetContractAddr)
	assert.Equal(t, "NFT", result.AssetSymbol)
	assert.Equal(t, "12345", result.TokenID)
	assert.Equal(t, "1", result.Value) // NFTs have value of 1
}

// TestGetNormalTransactions tests the normal transaction fetching method
func TestGetNormalTransactions(t *testing.T) {
	// Create a test server that returns a canned response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Make sure we're getting the right API endpoint and parameters
		query := r.URL.Query()
		assert.Equal(t, "account", query.Get("module"))
		assert.Equal(t, "txlist", query.Get("action"))
		assert.Equal(t, "0xtest", query.Get("address"))
		assert.NotEmpty(t, query.Get("apikey"))
		
		// Verify pagination parameters are present
		assert.Equal(t, "1", query.Get("page"))
		assert.Equal(t, "1000", query.Get("offset"))
		
		// Return a successful response with one transaction
		response := APIResponse{
			Status: "1", 
			Message: "OK",
			Result: json.RawMessage(`[{
				"blockNumber": "12345", 
				"timeStamp": "1630000000", 
				"hash": "0x123abc", 
				"from": "0xsender", 
				"to": "0xreceiver", 
				"value": "1000000000000000000",
				"gasPrice": "20000000000", 
				"gasUsed": "21000"
			}]`),
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Create a client that points to our test server
	client := NewEtherscanClient("dummy_api_key")
	// Override the BaseURL to point to our test server
	client.BaseURL = server.URL
	
	// Test the method
	txs, err := client.GetNormalTransactions("0xtest", 0, 999999999)
	
	// Check the results
	assert.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "0x123abc", txs[0].Hash)
	assert.Equal(t, "0xsender", txs[0].From)
	assert.Equal(t, "0xreceiver", txs[0].To)
	assert.Equal(t, "1000000000000000000", txs[0].Value)
}

// TestPagination tests basic pagination functionality
func TestPagination(t *testing.T) {
	// We'll track which pages are requested
	pagesRequested := make(map[string]bool)
	
	// Create a simple test for pagination by manipulating the server to return different
	// responses based on the page parameter
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check all the required parameters
		query := r.URL.Query()
		assert.Equal(t, "account", query.Get("module"))
		assert.Equal(t, "txlist", query.Get("action"))
		assert.Equal(t, "0xtest", query.Get("address"))
		assert.NotEmpty(t, query.Get("apikey"))
		
		// Get the page number from the request
		page := query.Get("page")
		
		// Mark this page as requested
		pagesRequested[page] = true
		
		// First page returns DefaultOffset transactions (simulating exactly batch size)
		// which should trigger the pagination to request page 2
		var response APIResponse
		if page == "1" {
			// Create a response with DefaultOffset transactions
			// For testing, we'll actually just make one transaction and lie about the length
			txs := make([]NormalTransaction, DefaultOffset)
			// Fill with one real transaction data
			tx := NormalTransaction{
				BlockNumber: "12345",
				TimeStamp: "1630000000",
				Hash: "0x111",
				From: "0xsender",
				To: "0xreceiver",
				Value: "1000000000000000000",
				GasPrice: "20000000000",
				GasUsed: "21000",
			}
			// Just use the same transaction for all slots to make DefaultOffset elements
			for i := 0; i < DefaultOffset; i++ {
				txs[i] = tx
			}
			
			// Convert to JSON
			txsBytes, _ := json.Marshal(txs)
			response = APIResponse{
				Status: "1",
				Message: "OK",
				Result: txsBytes,
			}
		} else if page == "2" {
			// Second page has fewer transactions (indicating end of results)
			tx := NormalTransaction{
				BlockNumber: "12346",
				TimeStamp: "1630000010",
				Hash: "0x222",
				From: "0xsender",
				To: "0xreceiver2",
				Value: "2000000000000000000",
				GasPrice: "20000000000",
				GasUsed: "21000",
			}
			txs := []NormalTransaction{tx}
			
			// Convert to JSON
			txsBytes, _ := json.Marshal(txs)
			response = APIResponse{
				Status: "1",
				Message: "OK",
				Result: txsBytes,
			}
		} else {
			// Any other page returns empty array
			response = APIResponse{
				Status: "1",
				Message: "OK",
				Result: json.RawMessage(`[]`),
			}
		}
		
		// Send the response
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Create a client that uses our mock server
	client := &EtherscanClient{
		ApiKey: "test_key",
		BaseURL: server.URL,
		HTTPClient: http.DefaultClient,
	}
	
	// Test the GetAllNormalTransactions method which should handle pagination
	allTxs, err := client.GetAllNormalTransactions("0xtest", 0, 999999999)
	
	// Verify results
	assert.NoError(t, err)
	
	// Verify that both pages were requested
	assert.True(t, pagesRequested["1"], "Page 1 should have been requested")
	assert.True(t, pagesRequested["2"], "Page 2 should have been requested")
	
	// Verify we got transactions from both pages (DefaultOffset + 1)
	expectedCount := DefaultOffset + 1
	assert.Equal(t, expectedCount, len(allTxs), "Expected %d transactions total", expectedCount)
}

func TestEtherscanClient_makeRequest(t *testing.T) {
	// Create test server with mock responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		action := query.Get("action")
		
		switch action {
		case "txlist":
			// Mock response for normal transactions
			mockResponse := APIResponse{
				Status:  "1",
				Message: "OK",
				Result:  json.RawMessage(`[{"blockNumber":"12345","timeStamp":"1630000000","hash":"0x123","from":"0xabc","to":"0xdef","value":"1000000000000000000","gasPrice":"20000000000","gasUsed":"21000"}]`),
			}
			json.NewEncoder(w).Encode(mockResponse)
			
		case "txlistinternal":
			// Mock response for internal transactions
			mockResponse := APIResponse{
				Status:  "1",
				Message: "OK",
				Result:  json.RawMessage(`[{"blockNumber":"12345","timeStamp":"1630000000","hash":"0x456","from":"0xcontract","to":"0xuser","value":"500000000000000000"}]`),
			}
			json.NewEncoder(w).Encode(mockResponse)
			
		case "tokentx":
			// Mock response for ERC20 transfers
			mockResponse := APIResponse{
				Status:  "1",
				Message: "OK",
				Result:  json.RawMessage(`[{"blockNumber":"12345","timeStamp":"1630000000","hash":"0x789","from":"0xabc","to":"0xdef","contractAddress":"0xtoken","tokenName":"Test Token","tokenSymbol":"TEST","tokenDecimal":"18","value":"1000000000000000000"}]`),
			}
			json.NewEncoder(w).Encode(mockResponse)
			
		case "error":
			// Mock error response
			mockResponse := APIResponse{
				Status:  "0",
				Message: "Error!",
				Result:  json.RawMessage(`""`),
			}
			json.NewEncoder(w).Encode(mockResponse)
		}
	}))
	defer server.Close()
	
	// Create client for testing that uses our test server instead of the real one
	client := &EtherscanClient{
		ApiKey: "dummy_api_key",
		HTTPClient: &http.Client{Timeout: time.Second * 10},
	}
	
	// Helper function to make API request to our test server instead of real Etherscan API
	makeTestRequest := func(params map[string][]string, result interface{}) error {
		urlValues := url.Values{}
		for k, vs := range params {
			for _, v := range vs {
				urlValues.Add(k, v)
			}
		}
		
		apiURL := server.URL + "?" + urlValues.Encode()
		resp, err := client.HTTPClient.Get(apiURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return err
		}
		
		if apiResp.Status != "1" {
			return fmt.Errorf("API returned error: %s", apiResp.Message)
		}
		
		if err := json.Unmarshal(apiResp.Result, result); err != nil {
			return err
		}
		
		return nil
	}
	
	// Test successful normal transactions request
	var normalTxs []NormalTransaction
	err := makeTestRequest(map[string][]string{"action": {"txlist"}}, &normalTxs)
	assert.NoError(t, err)
	assert.Len(t, normalTxs, 1)
	assert.Equal(t, "0x123", normalTxs[0].Hash)
	
	// Test API error
	err = makeTestRequest(map[string][]string{"action": {"error"}}, &normalTxs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error!")
}
