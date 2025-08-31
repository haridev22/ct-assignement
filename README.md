# Ethereum Transaction History Exporter

A Go application that retrieves transaction history for a specified Ethereum wallet address and exports it to a structured CSV file. This tool captures various transaction types including:

- Normal ETH transfers
- Internal transfers
- ERC-20 token transfers
- ERC-721 NFT transfers

## Features

- Fetches comprehensive transaction history for any Ethereum address
- Processes and categorizes different transaction types
- Exports transaction data to a CSV file with detailed information
- Handles concurrent API requests for faster data retrieval
- Supports block range filtering
- Implements pagination for handling large transaction volumes
- Provides automatic retry with exponential backoff for API rate limits
- Supports batch processing for large addresses

## Requirements

- Go 1.16 or later
- Etherscan API key (get one at https://etherscan.io/myapikey)

## Installation

1. Clone the repository or download the source code
2. Navigate to the project directory
3. Build the application:

```bash
go mod tidy
go build -o eth-tx-exporter
```

## Usage

```bash
./eth-tx-exporter -address 0xYourEthereumAddress -apikey YourEtherscanAPIKey
```

### Command Line Options

- `-address` (required): The Ethereum wallet address to fetch transactions for
- `-apikey` (required): Your Etherscan API key (can also be set via ETHERSCAN_API_KEY environment variable)
- `-output` (optional): Directory to save CSV output (default: "./output")
- `-start` (optional): Starting block number (default: 0)
- `-end` (optional): Ending block number (default: 999999999)
- `-batch` (optional): Process in smaller block ranges (e.g., 100000 blocks at a time)

### Example

```bash
./eth-tx-exporter -address 0xa39b189482f984388a34460636fea9eb181ad1a6 -apikey ABC123DEF456
```

## Output

The application generates a CSV file with the following fields:

- Transaction Hash
- Date & Time
- From Address
- To Address
- Transaction Type (ETH_TRANSFER, ERC20_TRANSFER, ERC721_TRANSFER, INTERNAL_TRANSFER, etc.)
- Asset Contract Address (if applicable)
- Asset Symbol / Name (if applicable)
- Token ID (for NFTs)
- Value / Amount
- Gas Fee (in ETH)

Output files are named using the format: `[address]_tx_history.csv`

When using batch processing, intermediate files will be saved with the format: `[address]_tx_history_blocks_[startBlock]_[endBlock].csv` and a final combined file as `[address]_tx_history_full.csv`

## Performance Considerations

For addresses with a large number of transactions, the application provides several features to handle them efficiently:

1. **Batch Processing**: Use the `-batch` flag to process transactions in smaller block ranges. For example:
   ```bash
   ./eth-tx-exporter -address 0xYourAddress -apikey YourAPIKey -batch 100000
   ```
   This will process transactions in chunks of 100,000 blocks at a time, which helps with memory usage and provides intermediate results.

2. **Pagination**: The application automatically handles pagination for API responses that exceed the maximum records per request (1,000).

3. **Retry Logic**: Built-in retry mechanism with exponential backoff handles API rate limiting and transient errors gracefully.

4. **Block Range Filtering**: For targeted analysis, you can specify precise block ranges with `-start` and `-end` flags.

## Assumptions

The following assumptions were made during the development of this project:

1. **API Limitations**: Etherscan API has rate limits and pagination constraints (max 1,000 records per request). The implementation assumes these limitations will remain consistent.

2. **Transaction Types**: The project assumes that normal, internal, ERC-20, and ERC-721 transactions cover the majority of relevant transaction types for most use cases. Other transaction types like ERC-1155 could be added in future versions.

3. **Block Finality**: The exporter assumes that block data beyond a certain age is final and won't be subject to reorgs, so repeated exports with the same parameters should yield consistent results.

4. **Data Availability**: The project assumes that Etherscan's API provides complete and accurate transaction history, which may not always be the case for very old transactions or during network congestion.

5. **Address Format**: The exporter assumes that input addresses follow the standard Ethereum address format (0x followed by 40 hexadecimal characters) but does not perform extensive validation.

6. **CSV as Export Format**: The project assumes CSV is an adequate format for most users' export needs. More complex data formats could be supported in future versions.

## Architecture Decisions

1. **Modular Design**: The codebase is organized into separate packages (api, models, utils) to enhance maintainability and support future extensions.

2. **Concurrent Processing**: Multiple transaction types are fetched concurrently using goroutines and wait groups, significantly improving performance compared to sequential processing.

3. **Batch Processing with Intermediate Results**: For addresses with massive transaction volumes, the batch processing approach not only manages memory more effectively but also provides users with intermediate results.

4. **Unified Transaction Model**: All transaction types are converted to a common transaction model, making the export process uniform regardless of the original transaction type.

5. **Exponential Backoff for Retries**: The implementation uses exponential backoff for API request retries, balancing between quickly recovering from transient errors and not overwhelming the API during persistent issues.

6. **Paginated Data Retrieval**: Instead of assuming a maximum transaction count, the exporter implements proper pagination to handle addresses with any number of transactions.

7. **In-Memory Processing**: The current implementation processes transactions in memory before writing to CSV, which prioritizes performance for typical use cases over handling extreme edge cases with millions of transactions.

8. **Direct API Integration**: Rather than using a third-party Ethereum client library, the project directly integrates with Etherscan's API for better control over requests and response handling.

