package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"eth-tx-history/pkg/models"
)

const (
	// EtherscanBaseURL is the base URL for Etherscan API
	EtherscanBaseURL = "https://api.etherscan.io/api"
)

// EtherscanClient represents an Etherscan API client
type EtherscanClient struct {
	ApiKey     string
	BaseURL    string
	MaxRetries int
	RetryDelay time.Duration
	HTTPClient *http.Client
}

// NewEtherscanClient creates a new Etherscan API client
func NewEtherscanClient(apiKey string) *EtherscanClient {
	return &EtherscanClient{
		ApiKey:     apiKey,
		BaseURL:    EtherscanBaseURL,
		MaxRetries: 3,
		RetryDelay: time.Second * 1,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// NormalTransaction represents a normal ETH transaction from Etherscan API
type NormalTransaction struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	IsError           string `json:"isError"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
}

// InternalTransaction represents an internal transaction from Etherscan API
type InternalTransaction struct {
	BlockNumber     string `json:"blockNumber"`
	TimeStamp       string `json:"timeStamp"`
	Hash            string `json:"hash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	ContractAddress string `json:"contractAddress"`
	Type            string `json:"type"`
	IsError         string `json:"isError"`
}

// ERC20Transaction represents an ERC20 token transfer from Etherscan API
type ERC20Transaction struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	ContractAddress   string `json:"contractAddress"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
}

// ERC721Transaction represents an ERC721 NFT transfer from Etherscan API
type ERC721Transaction struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	From              string `json:"from"`
	To                string `json:"to"`
	TokenID           string `json:"tokenID"`
	ContractAddress   string `json:"contractAddress"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
}

// APIResponse represents the response from Etherscan API
type APIResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// Default pagination settings
const (
	DefaultPage   = 1
	DefaultOffset = 1000 // Max allowed by Etherscan API
)

// GetNormalTransactions fetches normal transactions for the given address with pagination
func (c *EtherscanClient) GetNormalTransactions(address string, startBlock, endBlock int64) ([]NormalTransaction, error) {
	return c.GetNormalTransactionsPaginated(address, startBlock, endBlock, DefaultPage, DefaultOffset)
}

// GetNormalTransactionsPaginated fetches normal transactions for the given address with pagination
func (c *EtherscanClient) GetNormalTransactionsPaginated(address string, startBlock, endBlock int64, page, offset int) ([]NormalTransaction, error) {
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "txlist")
	params.Add("address", address)
	params.Add("startblock", strconv.FormatInt(startBlock, 10))
	params.Add("endblock", strconv.FormatInt(endBlock, 10))
	params.Add("page", strconv.Itoa(page))
	params.Add("offset", strconv.Itoa(offset))
	params.Add("sort", "asc")
	params.Add("apikey", c.ApiKey)

	var transactions []NormalTransaction
	if err := c.requestWithRetry(params, &transactions); err != nil {
		return nil, err
	}
	
	// Log progress if not empty
	if len(transactions) > 0 {
		fmt.Printf("Fetched %d normal transactions (page %d)\n", len(transactions), page)
	}
	return transactions, nil
}

// GetAllNormalTransactions fetches all normal transactions for the given address using pagination
func (c *EtherscanClient) GetAllNormalTransactions(address string, startBlock, endBlock int64) ([]NormalTransaction, error) {
	var allTransactions []NormalTransaction
	page := 1
	batchSize := DefaultOffset

	for {
		fmt.Printf("Fetching normal transactions page %d...\n", page)
		transactions, err := c.GetNormalTransactionsPaginated(address, startBlock, endBlock, page, batchSize)
		if err != nil {
			return nil, err
		}
		
		allTransactions = append(allTransactions, transactions...)
		
		// If we got fewer results than the batch size, we've reached the end
		if len(transactions) < batchSize {
			break
		}
		
		page++
		// Add a small delay between requests to avoid rate limits
		time.Sleep(200 * time.Millisecond)
	}
	
	fmt.Printf("Total normal transactions fetched: %d\n", len(allTransactions))
	return allTransactions, nil
}

// GetInternalTransactions fetches internal transactions for the given address
func (c *EtherscanClient) GetInternalTransactions(address string, startBlock, endBlock int64) ([]InternalTransaction, error) {
	return c.GetInternalTransactionsPaginated(address, startBlock, endBlock, DefaultPage, DefaultOffset)
}

// GetInternalTransactionsPaginated fetches internal transactions for the given address with pagination
func (c *EtherscanClient) GetInternalTransactionsPaginated(address string, startBlock, endBlock int64, page, offset int) ([]InternalTransaction, error) {
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "txlistinternal")
	params.Add("address", address)
	params.Add("startblock", strconv.FormatInt(startBlock, 10))
	params.Add("endblock", strconv.FormatInt(endBlock, 10))
	params.Add("page", strconv.Itoa(page))
	params.Add("offset", strconv.Itoa(offset))
	params.Add("sort", "asc")
	params.Add("apikey", c.ApiKey)

	var transactions []InternalTransaction
	if err := c.requestWithRetry(params, &transactions); err != nil {
		return nil, err
	}
	
	// Log progress if not empty
	if len(transactions) > 0 {
		fmt.Printf("Fetched %d internal transactions (page %d)\n", len(transactions), page)
	}
	return transactions, nil
}

// GetAllInternalTransactions fetches all internal transactions for the given address using pagination
func (c *EtherscanClient) GetAllInternalTransactions(address string, startBlock, endBlock int64) ([]InternalTransaction, error) {
	var allTransactions []InternalTransaction
	page := 1
	batchSize := DefaultOffset

	for {
		fmt.Printf("Fetching internal transactions page %d...\n", page)
		transactions, err := c.GetInternalTransactionsPaginated(address, startBlock, endBlock, page, batchSize)
		if err != nil {
			return nil, err
		}
		
		allTransactions = append(allTransactions, transactions...)
		
		// If we got fewer results than the batch size, we've reached the end
		if len(transactions) < batchSize {
			break
		}
		
		page++
		// Add a small delay between requests to avoid rate limits
		time.Sleep(200 * time.Millisecond)
	}
	
	fmt.Printf("Total internal transactions fetched: %d\n", len(allTransactions))
	return allTransactions, nil
}

// GetERC20Transfers fetches ERC20 token transfers for the given address
func (c *EtherscanClient) GetERC20Transfers(address string, startBlock, endBlock int64) ([]ERC20Transaction, error) {
	return c.GetERC20TransfersPaginated(address, startBlock, endBlock, DefaultPage, DefaultOffset)
}

// GetERC20TransfersPaginated fetches ERC20 token transfers for the given address with pagination
func (c *EtherscanClient) GetERC20TransfersPaginated(address string, startBlock, endBlock int64, page, offset int) ([]ERC20Transaction, error) {
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "tokentx")
	params.Add("address", address)
	params.Add("startblock", strconv.FormatInt(startBlock, 10))
	params.Add("endblock", strconv.FormatInt(endBlock, 10))
	params.Add("page", strconv.Itoa(page))
	params.Add("offset", strconv.Itoa(offset))
	params.Add("sort", "asc")
	params.Add("apikey", c.ApiKey)

	var transactions []ERC20Transaction
	if err := c.requestWithRetry(params, &transactions); err != nil {
		return nil, err
	}
	
	// Log progress if not empty
	if len(transactions) > 0 {
		fmt.Printf("Fetched %d ERC20 token transfers (page %d)\n", len(transactions), page)
	}
	return transactions, nil
}

// GetAllERC20Transfers fetches all ERC20 token transfers for the given address using pagination
func (c *EtherscanClient) GetAllERC20Transfers(address string, startBlock, endBlock int64) ([]ERC20Transaction, error) {
	var allTransactions []ERC20Transaction
	page := 1
	batchSize := DefaultOffset

	for {
		fmt.Printf("Fetching ERC20 token transfers page %d...\n", page)
		transactions, err := c.GetERC20TransfersPaginated(address, startBlock, endBlock, page, batchSize)
		if err != nil {
			return nil, err
		}
		
		allTransactions = append(allTransactions, transactions...)
		
		// If we got fewer results than the batch size, we've reached the end
		if len(transactions) < batchSize {
			break
		}
		
		page++
		// Add a small delay between requests to avoid rate limits
		time.Sleep(200 * time.Millisecond)
	}
	
	fmt.Printf("Total ERC20 token transfers fetched: %d\n", len(allTransactions))
	return allTransactions, nil
}

// GetERC721Transfers fetches ERC721 NFT transfers for the given address
func (c *EtherscanClient) GetERC721Transfers(address string, startBlock, endBlock int64) ([]ERC721Transaction, error) {
	return c.GetERC721TransfersPaginated(address, startBlock, endBlock, DefaultPage, DefaultOffset)
}

// GetERC721TransfersPaginated fetches ERC721 NFT transfers for the given address with pagination
func (c *EtherscanClient) GetERC721TransfersPaginated(address string, startBlock, endBlock int64, page, offset int) ([]ERC721Transaction, error) {
	params := url.Values{}
	params.Add("module", "account")
	params.Add("action", "tokennfttx")
	params.Add("address", address)
	params.Add("startblock", strconv.FormatInt(startBlock, 10))
	params.Add("endblock", strconv.FormatInt(endBlock, 10))
	params.Add("page", strconv.Itoa(page))
	params.Add("offset", strconv.Itoa(offset))
	params.Add("sort", "asc")
	params.Add("apikey", c.ApiKey)

	var transactions []ERC721Transaction
	if err := c.requestWithRetry(params, &transactions); err != nil {
		return nil, err
	}
	
	// Log progress if not empty
	if len(transactions) > 0 {
		fmt.Printf("Fetched %d ERC721 NFT transfers (page %d)\n", len(transactions), page)
	}
	return transactions, nil
}

// GetAllERC721Transfers fetches all ERC721 NFT transfers for the given address using pagination
func (c *EtherscanClient) GetAllERC721Transfers(address string, startBlock, endBlock int64) ([]ERC721Transaction, error) {
	var allTransactions []ERC721Transaction
	page := 1
	batchSize := DefaultOffset

	for {
		fmt.Printf("Fetching ERC721 NFT transfers page %d...\n", page)
		transactions, err := c.GetERC721TransfersPaginated(address, startBlock, endBlock, page, batchSize)
		if err != nil {
			return nil, err
		}
		
		allTransactions = append(allTransactions, transactions...)
		
		// If we got fewer results than the batch size, we've reached the end
		if len(transactions) < batchSize {
			break
		}
		
		page++
		// Add a small delay between requests to avoid rate limits
		time.Sleep(200 * time.Millisecond)
	}
	
	fmt.Printf("Total ERC721 NFT transfers fetched: %d\n", len(allTransactions))
	return allTransactions, nil
}

// makeRequest makes an HTTP request to the Etherscan API with retries and exponential backoff
func (c *EtherscanClient) makeRequest(url string) ([]byte, error) {
	var resp *http.Response
	var err error
	var body []byte
	retries := 0
	delay := c.RetryDelay

	for retries <= c.MaxRetries {
		resp, err = c.HTTPClient.Get(url)
		if err != nil {
			retries++
			if retries > c.MaxRetries {
				return nil, err
			}
			fmt.Printf("Request failed (attempt %d/%d): %s. Retrying in %v...\n", 
				retries, c.MaxRetries, err.Error(), delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		// Check if we hit rate limits (status code 429) or other server errors (5xx)
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			retries++
			if retries > c.MaxRetries {
				return nil, fmt.Errorf("API request failed with status code: %d after %d retries", 
					resp.StatusCode, retries-1)
			}
			fmt.Printf("Rate limit hit or server error (attempt %d/%d): status %d. Retrying in %v...\n", 
				retries, c.MaxRetries, resp.StatusCode, delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	return nil, fmt.Errorf("failed to make API request after %d retries", c.MaxRetries)
}

// requestWithRetry makes a request to the Etherscan API with retries and exponential backoff
func (c *EtherscanClient) requestWithRetry(params url.Values, result interface{}) error {
	apiURL := fmt.Sprintf("%s?%s", c.BaseURL, params.Encode())
	body, err := c.makeRequest(apiURL)
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

// ConvertNormalTxToModel converts a normal transaction to a generic transaction model
func ConvertNormalTxToModel(tx NormalTransaction) (models.Transaction, error) {
	timestamp, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
	if err != nil {
		return models.Transaction{}, err
	}

	// Calculate gas fee
	gasPrice, _ := new(big.Int).SetString(tx.GasPrice, 10)
	gasUsed, _ := new(big.Int).SetString(tx.GasUsed, 10)
	gasFee := new(big.Int).Mul(gasPrice, gasUsed)
	
	// Convert wei to ETH (1 ETH = 10^18 wei)
	weiPerEth := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	gasFeeEth := new(big.Float).Quo(new(big.Float).SetInt(gasFee), weiPerEth)
	
	// Format to 18 decimal places
	gasFeeStr := gasFeeEth.Text('f', 18)
	
	// Convert wei value to ETH
	valueWei, _ := new(big.Int).SetString(tx.Value, 10)
	valueEth := new(big.Float).Quo(new(big.Float).SetInt(valueWei), weiPerEth)
	valueStr := valueEth.Text('f', 18)

	return models.Transaction{
		Hash:      tx.Hash,
		Timestamp: time.Unix(timestamp, 0),
		From:      tx.From,
		To:        tx.To,
		Type:      models.TypeEthTransfer,
		Value:     valueStr,
		GasFee:    gasFeeStr,
	}, nil
}

// ConvertInternalTxToModel converts an internal transaction to a generic transaction model
func ConvertInternalTxToModel(tx InternalTransaction) (models.Transaction, error) {
	timestamp, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
	if err != nil {
		return models.Transaction{}, err
	}

	// Convert wei value to ETH
	valueWei, _ := new(big.Int).SetString(tx.Value, 10)
	weiPerEth := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	valueEth := new(big.Float).Quo(new(big.Float).SetInt(valueWei), weiPerEth)
	valueStr := valueEth.Text('f', 18)

	return models.Transaction{
		Hash:      tx.Hash,
		Timestamp: time.Unix(timestamp, 0),
		From:      tx.From,
		To:        tx.To,
		Type:      models.TypeInternalTx,
		Value:     valueStr,
		GasFee:    "0", // Gas fees are paid by the parent transaction
	}, nil
}

// ConvertERC20TxToModel converts an ERC20 transaction to a generic transaction model
func ConvertERC20TxToModel(tx ERC20Transaction) (models.Transaction, error) {
	timestamp, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
	if err != nil {
		return models.Transaction{}, err
	}

	// Calculate gas fee
	gasPrice, _ := new(big.Int).SetString(tx.GasPrice, 10)
	gasUsed, _ := new(big.Int).SetString(tx.GasUsed, 10)
	gasFee := new(big.Int).Mul(gasPrice, gasUsed)
	
	// Convert wei to ETH for gas fee
	weiPerEth := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	gasFeeEth := new(big.Float).Quo(new(big.Float).SetInt(gasFee), weiPerEth)
	gasFeeStr := gasFeeEth.Text('f', 18)

	// Convert token value based on decimals
	tokenDecimals, _ := strconv.Atoi(tx.TokenDecimal)
	tokenValue, _ := new(big.Int).SetString(tx.Value, 10)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil))
	actualValue := new(big.Float).Quo(new(big.Float).SetInt(tokenValue), divisor)
	valueStr := actualValue.Text('f', tokenDecimals)

	return models.Transaction{
		Hash:              tx.Hash,
		Timestamp:         time.Unix(timestamp, 0),
		From:              tx.From,
		To:                tx.To,
		Type:              models.TypeERC20Transfer,
		AssetContractAddr: tx.ContractAddress,
		AssetSymbol:       tx.TokenSymbol,
		Value:             valueStr,
		GasFee:            gasFeeStr,
	}, nil
}

// ConvertERC721TxToModel converts an ERC721 transaction to a generic transaction model
func ConvertERC721TxToModel(tx ERC721Transaction) (models.Transaction, error) {
	timestamp, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
	if err != nil {
		return models.Transaction{}, err
	}

	// Calculate gas fee
	gasPrice, _ := new(big.Int).SetString(tx.GasPrice, 10)
	gasUsed, _ := new(big.Int).SetString(tx.GasUsed, 10)
	gasFee := new(big.Int).Mul(gasPrice, gasUsed)
	
	// Convert wei to ETH for gas fee
	weiPerEth := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	gasFeeEth := new(big.Float).Quo(new(big.Float).SetInt(gasFee), weiPerEth)
	gasFeeStr := gasFeeEth.Text('f', 18)

	return models.Transaction{
		Hash:              tx.Hash,
		Timestamp:         time.Unix(timestamp, 0),
		From:              tx.From,
		To:                tx.To,
		Type:              models.TypeERC721Transfer,
		AssetContractAddr: tx.ContractAddress,
		AssetSymbol:       tx.TokenSymbol,
		TokenID:           tx.TokenID,
		Value:             "1", // NFTs have a quantity of 1
		GasFee:            gasFeeStr,
	}, nil
}
