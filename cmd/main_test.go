package main

import (
	"encoding/json"
	"testing"

	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"go.uber.org/zap"
)

func Test_TaskRequestPayload(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger: %v", err)
	}

	taskWorker := NewTaskWorker(logger)

	// Test case 1: Valid token ID (bitcoin)
	t.Run("Valid Bitcoin Price Request", func(t *testing.T) {
		taskRequest := &performerV1.TaskRequest{
			TaskId:  []byte("test-task-bitcoin"),
			Payload: []byte("bitcoin"),
		}

		// Validate the task
		err = taskWorker.ValidateTask(taskRequest)
		if err != nil {
			t.Errorf("ValidateTask failed for valid token: %v", err)
		}

		// Handle the task
		resp, err := taskWorker.HandleTask(taskRequest)
		if err != nil {
			t.Errorf("HandleTask failed: %v", err)
		}

		// Parse the response
		var priceData PriceResponse
		err = json.Unmarshal(resp.Result, &priceData)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Validate response fields
		if priceData.TokenID != "bitcoin" {
			t.Errorf("Expected token ID 'bitcoin', got '%s'", priceData.TokenID)
		}
		if priceData.TokenName == "" {
			t.Error("Token name should not be empty")
		}
		if priceData.Price <= 0 {
			t.Error("Price should be greater than 0")
		}
		if priceData.Currency != "USD" {
			t.Errorf("Expected currency 'USD', got '%s'", priceData.Currency)
		}
		if priceData.Timestamp <= 0 {
			t.Error("Timestamp should be greater than 0")
		}

		t.Logf("Bitcoin price response: %+v", priceData)
	})

	// Test case 2: Valid token ID (ethereum)
	t.Run("Valid Ethereum Price Request", func(t *testing.T) {
		taskRequest := &performerV1.TaskRequest{
			TaskId:  []byte("test-task-ethereum"),
			Payload: []byte("ethereum"),
		}

		err = taskWorker.ValidateTask(taskRequest)
		if err != nil {
			t.Errorf("ValidateTask failed for ethereum: %v", err)
		}

		resp, err := taskWorker.HandleTask(taskRequest)
		if err != nil {
			t.Errorf("HandleTask failed for ethereum: %v", err)
		}

		var priceData PriceResponse
		err = json.Unmarshal(resp.Result, &priceData)
		if err != nil {
			t.Errorf("Failed to unmarshal ethereum response: %v", err)
		}

		if priceData.TokenID != "ethereum" {
			t.Errorf("Expected token ID 'ethereum', got '%s'", priceData.TokenID)
		}

		t.Logf("Ethereum price response: %+v", priceData)
	})

	// Test case 3: Empty payload validation should fail
	t.Run("Empty Payload Validation", func(t *testing.T) {
		taskRequest := &performerV1.TaskRequest{
			TaskId:  []byte("test-task-empty"),
			Payload: []byte(""),
		}

		err = taskWorker.ValidateTask(taskRequest)
		if err == nil {
			t.Error("ValidateTask should fail for empty payload")
		}
	})

	// Test case 4: Invalid token ID should fail
	t.Run("Invalid Token ID", func(t *testing.T) {
		taskRequest := &performerV1.TaskRequest{
			TaskId:  []byte("test-task-invalid"),
			Payload: []byte("invalid-token-id-xyz123"),
		}

		err = taskWorker.ValidateTask(taskRequest)
		if err != nil {
			t.Errorf("ValidateTask failed: %v", err)
		}

		// This should fail when trying to fetch from CoinGecko
		_, err = taskWorker.HandleTask(taskRequest)
		if err == nil {
			t.Error("HandleTask should fail for invalid token ID")
		}
	})
}

// Test the fetchTokenPrice method directly
func Test_FetchTokenPrice(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	taskWorker := NewTaskWorker(logger)

	// Test fetching bitcoin price
	t.Run("Fetch Bitcoin Price", func(t *testing.T) {
		priceData, err := taskWorker.fetchTokenPrice("bitcoin")
		if err != nil {
			t.Errorf("Failed to fetch bitcoin price: %v", err)
		}

		if priceData.TokenID != "bitcoin" {
			t.Errorf("Expected token ID 'bitcoin', got '%s'", priceData.TokenID)
		}
		if priceData.TokenName != "Bitcoin" {
			t.Errorf("Expected token name 'Bitcoin', got '%s'", priceData.TokenName)
		}
		if priceData.Price <= 0 {
			t.Error("Bitcoin price should be greater than 0")
		}

		t.Logf("Bitcoin price: $%.2f USD", priceData.Price)
	})

	// Test fetching invalid token
	t.Run("Fetch Invalid Token", func(t *testing.T) {
		_, err := taskWorker.fetchTokenPrice("invalid-token-xyz")
		if err == nil {
			t.Error("Should fail for invalid token")
		}
	})
}