package services

import (
	"fmt"
	"main/internal/models"
	"main/internal/repositories"
	"sync"
	"sync/atomic"
)

type OrderEventService interface {
	ImportOrderEvents(events []models.OrderEvent) (ImportOrderEventsResult, error)
}

type ImportOrderEventsResult struct {
	Total     int `json:"total"`
	Accepted  int `json:"accepted"`
	Rejected  int `json:"rejected"`
	Duplicate int `json:"duplicate"`
}

type orderEventService struct {
	eventRepo repositories.OrderEventRepository
}

func NewOrderEventService(eventRepo repositories.OrderEventRepository) OrderEventService {
	return &orderEventService{
		eventRepo: eventRepo,
	}
}

func (s *orderEventService) ImportOrderEvents(events []models.OrderEvent) (ImportOrderEventsResult, error) {
	const workerCount = 5

	jobs := make(chan models.OrderEvent)
	var wg sync.WaitGroup
	var duplicateCount atomic.Int64
	var rejectedCount atomic.Int64

	seen := sync.Map{}
	validEvents := make(chan models.OrderEvent)

	worker := func() {
		defer wg.Done()
		for event := range jobs {

			if _, loaded := seen.LoadOrStore(event.EventID, true); loaded {
				duplicateCount.Add(1)
				fmt.Println("duplicate event:", event.EventID)
				continue
			}

			if err := models.ValidateEvent(&event); err != nil {
				rejectedCount.Add(1)
				fmt.Println("invalid event:", err)
				continue
			}

			validEvents <- event
		}
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		for _, e := range events {
			jobs <- e
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(validEvents)
	}()

	var toInsert []models.OrderEvent
	for e := range validEvents {
		toInsert = append(toInsert, e)
	}

	result := ImportOrderEventsResult{
		Total:     len(events),
		Accepted:  len(toInsert),
		Rejected:  int(rejectedCount.Load()),
		Duplicate: int(duplicateCount.Load()),
	}

	if len(toInsert) == 0 {
		return result, nil
	}
	if err := s.eventRepo.ImportOrderEvents(toInsert); err != nil {
		return result, err
	}
	return result, nil
}
