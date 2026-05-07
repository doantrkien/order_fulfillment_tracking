package worker

import (
	"encoding/json"
	"fmt"
	"main/internal/models"
	"os"
	"testing"
)

func TestWorkerWithMockData(t *testing.T) {
	fileBytes, err := os.ReadFile("../../pkg/mock_data/input.json")
	if err != nil {
		t.Fatalf("Failed to read mock data: %v", err)
	}

	var events []models.OrderEvent
	if err := json.Unmarshal(fileBytes, &events); err != nil {
		t.Fatalf("Failed to parse mock data: %v", err)
	}

	dbOrders = make(map[int64]models.OrderStatus)
	dbSeenEvents = make(map[string]bool)

	ordersBytes, err := os.ReadFile("../../pkg/mock_data/orders.json")
	if err == nil {
		var existingOrders []models.Order
		if json.Unmarshal(ordersBytes, &existingOrders) == nil {
			for _, o := range existingOrders {
				dbOrders[o.ID] = o.Status
			}
		}
	}

	fmt.Printf("dbOrders before RunBatch: %+v\n", dbOrders)

	fmt.Println("Running worker with events from mock data...")
	RunBatch(events)
}
