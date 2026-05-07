package worker

import (
	"encoding/json"
	"fmt"
	"main/internal/models"
	"sync"
)

// --- 1. Core Structs ---

type BatchResult struct {
	AcceptedCount  int `json:"accepted_count"`
	RejectedCount  int `json:"rejected_count"`
	DuplicateCount int `json:"duplicate_count"`
}

// --- 2. In-Memory "Database" ---

var (
	dbOrders     = make(map[int64]models.OrderStatus) // Maps OrderID to its current Status
	dbSeenEvents = make(map[string]bool)              // Keeps track of EventIDs we've already processed
	dbMutex      sync.Mutex                           // The lock to prevent data crashes
)

// --- 3. Helper Functions ---

// isValidTransition checks our business rules for order states
func isValidTransition(currentStatus, newStatus models.OrderStatus) bool {
	// If the order doesn't exist yet, it MUST be "created"
	if currentStatus == "" {
		return newStatus == models.OrderStatusCreated
	}

	// If it does exist, check the rules
	switch currentStatus {
	case models.OrderStatusCreated:
		return newStatus == models.OrderStatusPaid || newStatus == models.OrderStatusCanceled
	case models.OrderStatusPaid:
		return newStatus == models.OrderStatusPacked || newStatus == models.OrderStatusRefunded
	case models.OrderStatusPacked:
		return newStatus == models.OrderStatusShipped
	case models.OrderStatusShipped:
		return newStatus == models.OrderStatusDelivered
	default:
		return false // delivered, canceled, and refunded are final states
	}
}

// checkAndMarkDuplicate safely checks and updates our seen events list
func checkAndMarkDuplicate(eventID string) bool {
	dbMutex.Lock()
	defer dbMutex.Unlock() // Ensures the lock is released when the function finishes

	// If it is on our checklist, it is a duplicate
	if dbSeenEvents[eventID] {
		return true
	}

	// Otherwise, mark it as seen and return false (not a duplicate)
	dbSeenEvents[eventID] = true
	return false
}

// --- 4. The Worker ---

func Worker(jobs <-chan models.OrderEvent, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done() // Tell the WaitGroup we are done when this function finishes

	for event := range jobs {

		// 1. DUPLICATE CHECK
		// Use our dedicated function to safely check the event ID
		if checkAndMarkDuplicate(event.EventID) {
			results <- "duplicate"
			continue
		}

		// 2. STATE MACHINE LOGIC
		// Lock the door to safely check and update the order status
		dbMutex.Lock()

		currentStatus := dbOrders[event.OrderID]
		if !isValidTransition(currentStatus, event.NewStatus) {
			dbMutex.Unlock() // Must unlock before skipping to the next job
			results <- "rejected"
			continue
		}

		// If the transition is valid, update the order and unlock
		dbOrders[event.OrderID] = event.NewStatus
		dbMutex.Unlock()

		results <- "accepted"
	}
}

// --- 5. Main Execution Example ---

func RunBatch(events []models.OrderEvent) {
	// Setup channels
	jobs := make(chan models.OrderEvent, len(events))
	results := make(chan string, len(events))
	var wg sync.WaitGroup

	// Step 1: Start 3 concurrent workers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go Worker(jobs, results, &wg)
	}

	// Step 2: Feed the jobs channel
	for _, event := range events {
		jobs <- event
	}
	close(jobs) // Signal that no more jobs will be added

	// Step 3: Wait for all workers to finish, then close the results channel
	wg.Wait()
	close(results)

	// Step 4: Tally up the results
	var finalScore BatchResult
	for result := range results {
		switch result {
		case "accepted":
			finalScore.AcceptedCount++
		case "rejected":
			finalScore.RejectedCount++
		case "duplicate":
			finalScore.DuplicateCount++
		}
	}

	// Output the final result as JSON
	output, _ := json.Marshal(finalScore)
	fmt.Println(string(output))
}
