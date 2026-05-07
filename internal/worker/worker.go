package worker

import (
	"encoding/json"
	"fmt"
	"main/internal/models"
	"sync"
)


type BatchResult struct {
	AcceptedCount  int `json:"accepted_count"`
	RejectedCount  int `json:"rejected_count"`
	DuplicateCount int `json:"duplicate_count"`
}


var (
	dbOrders     = make(map[int64]models.OrderStatus) 
	dbSeenEvents = make(map[string]bool)              
	dbMutex      sync.Mutex                           
)


func isValidTransition(currentStatus, newStatus models.OrderStatus) bool {
	if currentStatus == "" {
		return newStatus == models.OrderStatusCreated
	}

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
		return false 
	}
}

func checkAndMarkDuplicate(eventID string) bool {
	dbMutex.Lock()
	defer dbMutex.Unlock() 
	if dbSeenEvents[eventID] {
		return true
	}

	dbSeenEvents[eventID] = true
	return false
}


func Worker(jobs <-chan models.OrderEvent, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done() 

	for event := range jobs {

		if checkAndMarkDuplicate(event.EventID) {
			results <- "duplicate"
			continue
		}

		dbMutex.Lock()

		currentStatus := dbOrders[event.OrderID]
		if !isValidTransition(currentStatus, event.NewStatus) {
			dbMutex.Unlock() 
			results <- "rejected"
			continue
		}

		dbOrders[event.OrderID] = event.NewStatus
		fmt.Printf("dbOrders after update (event_id=%s): %+v\n", event.EventID, dbOrders)
		dbMutex.Unlock()

		results <- "accepted"
	}
}


func RunBatch(events []models.OrderEvent) {
	jobs := make(chan models.OrderEvent, len(events))
	results := make(chan string, len(events))
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go Worker(jobs, results, &wg)
	}

	for _, event := range events {
		jobs <- event
	}
	close(jobs) 

	wg.Wait()
	close(results)

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

	output, _ := json.Marshal(finalScore)
	fmt.Println(string(output))
}
