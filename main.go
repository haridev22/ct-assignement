package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"eth-tx-history/pkg/api"
	"eth-tx-history/pkg/models"
	"eth-tx-history/pkg/utils"
)

const (
	// Default values
	defaultOutputDir      = "./output"
	defaultStartBlock     = 0
	defaultEndBlock       = 999999999 // to get all transactions
	maxConcurrentRequests = 4         // concurrent API requests
)

func main() {
	//command line flags
	address := flag.String("address", "", "Ethereum wallet address to fetch transactions for (required)")
	apiKey := flag.String("apikey", "", "Etherscan API key (required)")
	outputDir := flag.String("output", defaultOutputDir, "Directory to save CSV output")
	startBlock := flag.Int64("start", defaultStartBlock, "Starting block number")
	endBlock := flag.Int64("end", defaultEndBlock, "Ending block number")
	batchBlocks := flag.Int64("batch", 0, "Process in smaller block ranges (e.g., 100000 blocks at a time)")

	flag.Parse()

	if *address == "" {
		log.Fatal("Error: Ethereum wallet address is required. Use -address flag.")
	}

	// TODO: get api key from environment variable
	if *apiKey == "" {
		log.Fatal("Error: Etherscan API key is required. Use -apikey flag or set ETHERSCAN_API_KEY environment variable.")
	}

	client := api.NewEtherscanClient(*apiKey)

	fmt.Printf("Fetching transactions for address: %s\n", *address)
	fmt.Printf("Block range: %d to %d\n", *startBlock, *endBlock)

	// iif batch size specifiedthen process in batches
	if *batchBlocks > 0 {
		processInBatches(client, *address, *startBlock, *endBlock, *batchBlocks, *outputDir)
		return
	}

	var wg sync.WaitGroup
	wg.Add(4) // four transaction types

	// channel for transactions
	normalTxCh := make(chan []api.NormalTransaction, 1)
	internalTxCh := make(chan []api.InternalTransaction, 1)
	erc20TxCh := make(chan []api.ERC20Transaction, 1)
	erc721TxCh := make(chan []api.ERC721Transaction, 1)
	errorCh := make(chan error, 4)

	// Fetch normal ETH transactions with pagination
	go func() {
		defer wg.Done()
		fmt.Println("Starting to fetch normal ETH transactions...")
		txs, err := client.GetAllNormalTransactions(*address, *startBlock, *endBlock)
		if err != nil {
			errorCh <- fmt.Errorf("error fetching normal transactions: %w", err)
			normalTxCh <- nil
			return
		}
		normalTxCh <- txs
	}()

	// Fetch internal transactions with pagination
	go func() {
		defer wg.Done()
		fmt.Println("Starting to fetch internal transactions...")
		txs, err := client.GetAllInternalTransactions(*address, *startBlock, *endBlock)
		if err != nil {
			errorCh <- fmt.Errorf("error fetching internal transactions: %w", err)
			internalTxCh <- nil
			return
		}
		internalTxCh <- txs
	}()

	// Fetch ERC-20 token transfers with pagination
	go func() {
		defer wg.Done()
		fmt.Println("Starting to fetch ERC-20 token transfers...")
		txs, err := client.GetAllERC20Transfers(*address, *startBlock, *endBlock)
		if err != nil {
			errorCh <- fmt.Errorf("error fetching ERC-20 transfers: %w", err)
			erc20TxCh <- nil
			return
		}
		erc20TxCh <- txs
	}()

	// Fetch ERC-721 NFT transfers with pagination
	go func() {
		defer wg.Done()
		fmt.Println("Starting to fetch ERC-721 NFT transfers...")
		txs, err := client.GetAllERC721Transfers(*address, *startBlock, *endBlock)
		if err != nil {
			errorCh <- fmt.Errorf("error fetching ERC-721 transfers: %w", err)
			erc721TxCh <- nil
			return
		}
		erc721TxCh <- txs
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors
	select {
	case err := <-errorCh:
		log.Fatalf("Error: %v", err)
	default:
		// No errors
	}

	// Convert all transactions to a common model
	var allTxs []models.Transaction

	// normal transactions
	normalTxs := <-normalTxCh
	for _, tx := range normalTxs {
		model, err := api.ConvertNormalTxToModel(tx)
		if err != nil {
			log.Printf("Warning: Failed to process normal transaction %s: %v", tx.Hash, err)
			continue
		}
		allTxs = append(allTxs, model)
	}

	// internal transactions
	internalTxs := <-internalTxCh
	for _, tx := range internalTxs {
		model, err := api.ConvertInternalTxToModel(tx)
		if err != nil {
			log.Printf("Warning: Failed to process internal transaction %s: %v", tx.Hash, err)
			continue
		}
		allTxs = append(allTxs, model)
	}

	// ERC20 transactions
	erc20Txs := <-erc20TxCh
	for _, tx := range erc20Txs {
		model, err := api.ConvertERC20TxToModel(tx)
		if err != nil {
			log.Printf("Warning: Failed to process ERC20 transaction %s: %v", tx.Hash, err)
			continue
		}
		allTxs = append(allTxs, model)
	}

	// ERC721 transactions
	erc721Txs := <-erc721TxCh
	for _, tx := range erc721Txs {
		model, err := api.ConvertERC721TxToModel(tx)
		if err != nil {
			log.Printf("Warning: Failed to process ERC721 transaction %s: %v", tx.Hash, err)
			continue
		}
		allTxs = append(allTxs, model)
	}

	fmt.Printf("Total transactions processed: %d\n", len(allTxs))

	// Export to CSV
	fmt.Printf("Total transactions: %d\n", len(allTxs))

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Export to CSV
	filePath := filepath.Join(*outputDir, fmt.Sprintf("%s_tx_history.csv", *address))
	if err := utils.ExportTransactionsToCSV(allTxs, filePath); err != nil {
		log.Fatalf("Error exporting to CSV: %v", err)
	}

	fmt.Printf("Exported transaction history to %s\n", filePath)
}

// processInBatches processes transactions in smaller block ranges to avoid memory issues
func processInBatches(client *api.EtherscanClient, address string, startBlock, endBlock, batchSize int64, outputDir string) {
	var allTxs []models.Transaction
	var processedBlocks int64
	totalBlocks := endBlock - startBlock

	// Process in batches
	for currentStart := startBlock; currentStart < endBlock; currentStart += batchSize {
		currentEnd := currentStart + batchSize
		if currentEnd > endBlock {
			currentEnd = endBlock
		}

		fmt.Printf("\n=== Processing blocks %d to %d (%d%% complete) ===\n",
			currentStart, currentEnd, int(float64(processedBlocks)/float64(totalBlocks)*100))

		// Process each transaction type
		var batchTxs []models.Transaction

		// Normal transactions
		fmt.Println("Fetching normal transactions for batch...")
		normalTxs, err := client.GetAllNormalTransactions(address, currentStart, currentEnd)
		if err != nil {
			fmt.Printf("Warning: Error fetching normal transactions for block range %d-%d: %v\n",
				currentStart, currentEnd, err)
		} else {
			for _, tx := range normalTxs {
				convertedTx, err := api.ConvertNormalTxToModel(tx)
				if err == nil {
					batchTxs = append(batchTxs, convertedTx)
				}
			}
		}

		// Internal transactions
		fmt.Println("Fetching internal transactions for batch...")
		internalTxs, err := client.GetAllInternalTransactions(address, currentStart, currentEnd)
		if err != nil {
			fmt.Printf("Warning: Error fetching internal transactions for block range %d-%d: %v\n",
				currentStart, currentEnd, err)
		} else {
			for _, tx := range internalTxs {
				convertedTx, err := api.ConvertInternalTxToModel(tx)
				if err == nil {
					batchTxs = append(batchTxs, convertedTx)
				}
			}
		}

		// ERC20 transfers
		fmt.Println("Fetching ERC20 transfers for batch...")
		erc20Txs, err := client.GetAllERC20Transfers(address, currentStart, currentEnd)
		if err != nil {
			fmt.Printf("Warning: Error fetching ERC20 transfers for block range %d-%d: %v\n",
				currentStart, currentEnd, err)
		} else {
			for _, tx := range erc20Txs {
				convertedTx, err := api.ConvertERC20TxToModel(tx)
				if err == nil {
					batchTxs = append(batchTxs, convertedTx)
				}
			}
		}

		// ERC721 transfers
		fmt.Println("Fetching ERC721 transfers for batch...")
		erc721Txs, err := client.GetAllERC721Transfers(address, currentStart, currentEnd)
		if err != nil {
			fmt.Printf("Warning: Error fetching ERC721 transfers for block range %d-%d: %v\n",
				currentStart, currentEnd, err)
		} else {
			for _, tx := range erc721Txs {
				convertedTx, err := api.ConvertERC721TxToModel(tx)
				if err == nil {
					batchTxs = append(batchTxs, convertedTx)
				}
			}
		}

		// Append to all transactions
		allTxs = append(allTxs, batchTxs...)

		// Write intermediate results to CSV
		intermediateFilePath := filepath.Join(outputDir,
			fmt.Sprintf("%s_tx_history_blocks_%d_%d.csv", address, currentStart, currentEnd))
		if err := utils.ExportTransactionsToCSV(batchTxs, intermediateFilePath); err != nil {
			fmt.Printf("Warning: Error saving intermediate results: %v\n", err)
		} else {
			fmt.Printf("Saved intermediate results to %s\n", intermediateFilePath)
		}

		processedBlocks += (currentEnd - currentStart)
	}

	// Export final combined CSV
	finalFilePath := filepath.Join(outputDir, fmt.Sprintf("%s_tx_history_full.csv", address))
	if err := utils.ExportTransactionsToCSV(allTxs, finalFilePath); err != nil {
		log.Fatalf("Error exporting to CSV: %v", err)
	}

	fmt.Printf("\nComplete! Exported %d transactions to %s\n", len(allTxs), finalFilePath)
}
