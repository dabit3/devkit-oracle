package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer/server"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap"
)

// This offchain binary is run by Operators running the Hourglass Executor. It contains
// the business logic of the AVS and performs worked based on the tasked sent to it.
// The Hourglass Aggregator ingests tasks from the TaskMailbox and distributes work
// to Executors configured to run the AVS Performer. Performers execute the work and
// return the result to the Executor where the result is signed and return to the
// Aggregator to place in the outbox once the signing threshold is met.

type TaskWorker struct {
	logger *zap.Logger
}

// PriceResponse represents the structure we return to the caller
type PriceResponse struct {
	TokenID   string  `json:"token_id"`
	TokenName string  `json:"token_name"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Timestamp int64   `json:"timestamp"`
}

// CoinGeckoResponse represents the response from CoinGecko API
type CoinGeckoResponse struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	MarketData struct {
		CurrentPrice map[string]float64 `json:"current_price"`
	} `json:"market_data"`
}

func NewTaskWorker(logger *zap.Logger) *TaskWorker {
	return &TaskWorker{
		logger: logger,
	}
}

func (tw *TaskWorker) ValidateTask(t *performerV1.TaskRequest) error {
	tw.logger.Sugar().Infow("Validating task",
		zap.Any("task", t),
	)

	// Validate that payload is not empty
	if len(t.Payload) == 0 {
		return fmt.Errorf("task payload cannot be empty")
	}

	// Validate that payload is a valid string (token ID)
	tokenID := string(t.Payload)
	if tokenID == "" {
		return fmt.Errorf("token ID cannot be empty")
	}

	tw.logger.Sugar().Infow("Task validation successful",
		"tokenID", tokenID,
	)

	return nil
}

func (tw *TaskWorker) HandleTask(t *performerV1.TaskRequest) (*performerV1.TaskResponse, error) {
	tw.logger.Sugar().Infow("Handling task",
		zap.Any("task", t),
	)

	// Extract token ID from payload
	tokenID := string(t.Payload)
	tw.logger.Sugar().Infow("Fetching price for token", "tokenID", tokenID)

	// Call CoinGecko API to get token price
	priceData, err := tw.fetchTokenPrice(tokenID)
	if err != nil {
		tw.logger.Sugar().Errorw("Failed to fetch token price",
			"tokenID", tokenID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to fetch price for token %s: %w", tokenID, err)
	}

	// Log the price in human-readable format
	tw.logger.Sugar().Infow("Price fetched successfully",
		"token", priceData.TokenName,
		"tokenID", priceData.TokenID,
		"price", fmt.Sprintf("$%.2f", priceData.Price),
		"currency", priceData.Currency,
	)

	// Convert response to JSON
	resultBytes, err := json.Marshal(priceData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal price response: %w", err)
	}

	return &performerV1.TaskResponse{
		TaskId: t.TaskId,
		Result: resultBytes,
	}, nil
}

func (tw *TaskWorker) fetchTokenPrice(tokenID string) (*PriceResponse, error) {
	// CoinGecko API endpoint for token data
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false", tokenID)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make API request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call CoinGecko API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CoinGecko API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse CoinGecko response
	var cgResp CoinGeckoResponse
	if err := json.Unmarshal(body, &cgResp); err != nil {
		return nil, fmt.Errorf("failed to parse CoinGecko response: %w", err)
	}

	// Extract USD price
	usdPrice, ok := cgResp.MarketData.CurrentPrice["usd"]
	if !ok {
		return nil, fmt.Errorf("USD price not available for token %s", tokenID)
	}

	// Create price response
	priceResponse := &PriceResponse{
		TokenID:   cgResp.ID,
		TokenName: cgResp.Name,
		Price:     usdPrice,
		Currency:  "USD",
		Timestamp: time.Now().Unix(),
	}

	return priceResponse, nil
}

func main() {
	ctx := context.Background()
	l, _ := zap.NewProduction()

	w := NewTaskWorker(l)

	pp, err := server.NewPonosPerformerWithRpcServer(&server.PonosPerformerConfig{
		Port:    8080,
		Timeout: 5 * time.Second,
	}, w, l)
	if err != nil {
		panic(fmt.Errorf("failed to create performer: %w", err))
	}

	if err := pp.Start(ctx); err != nil {
		panic(err)
	}
}
