package services

import (
	"context"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"sort"
	"sync"
)

type OrderEventService interface {
	ImportOrderEvents(ctx context.Context, reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error)
}

type orderEventService struct {
	orderEventRepo repositories.OrderEventRepository
	maxWorkers     int
}

func NewOrderEventService(orderEventRepo repositories.OrderEventRepository, maxWorkers int) OrderEventService {
	return &orderEventService{
		orderEventRepo: orderEventRepo,
		maxWorkers:     maxWorkers,
	}
}

type batchResult struct {
	details []repositories.ProcessResultDetail
	err     error
}

func (s *orderEventService) ImportOrderEvents(ctx context.Context, reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error) {
	resp := dto.ImportOrderEventsResponse{
		Errors: []dto.EventError{},
	}

	var validReqs []dto.ImportOrderEventRequest

	for _, req := range reqs {
		if reason := validateBasic(req); reason != "" {
			resp.Rejected++
			resp.Errors = append(resp.Errors, dto.EventError{
				OrderID: req.OrderID,
				Status:  req.Status,
				Reason:  reason,
			})
			continue
		}
		validReqs = append(validReqs, req)
	}

	if len(validReqs) == 0 {
		return resp, nil
	}

	orderGroups := make(map[int64][]dto.ImportOrderEventRequest)
	for _, req := range validReqs {
		orderGroups[req.OrderID] = append(orderGroups[req.OrderID], req)
	}
	for _, events := range orderGroups {
		sort.Slice(events, func(i, j int) bool {
			return events[i].EventAt.Before(events[j].EventAt)
		})
	}

	orderedEvents := make([]models.OrderEvent, 0, len(validReqs))
	for _, events := range orderGroups {
		for _, req := range events {
			orderedEvents = append(orderedEvents, models.OrderEvent{
				OrderID:   req.OrderID,
				NewStatus: models.OrderStatus(req.Status),
				UpdatedBy: req.UpdatedBy,
				EventAt:   req.EventAt,
			})
		}
	}

	// Split into chunks for parallel processing by workers
	chunks := splitIntoChunks(orderedEvents, s.maxWorkers)

	var wg sync.WaitGroup
	resultsChan := make(chan batchResult, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(events []models.OrderEvent) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				resultsChan <- batchResult{err: ctx.Err()}
				return
			default:
			}
			details, err := s.orderEventRepo.ProcessBatchEventsTx(ctx, events)
			resultsChan <- batchResult{details: details, err: err}
		}(chunk.events)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var processingErr error
	for br := range resultsChan {
		if br.err != nil {
			processingErr = br.err
			continue
		}
		for _, detail := range br.details {
			switch detail.Result {
			case repositories.Accepted:
				resp.Accepted++
			case repositories.Rejected:
				resp.Rejected++
				resp.Errors = append(resp.Errors, dto.EventError{
					OrderID: detail.OrderID,
					Status:  detail.Status,
					Reason:  detail.Reason,
				})
			case repositories.Duplicate:
				resp.Duplicate++
				resp.Errors = append(resp.Errors, dto.EventError{
					OrderID: detail.OrderID,
					Status:  detail.Status,
					Reason:  detail.Reason,
				})
			}
		}
	}

	return resp, processingErr
}

// chunk holds a slice of events and their corresponding original requests
type chunk struct {
	events []models.OrderEvent
}

func splitIntoChunks(events []models.OrderEvent, numChunks int) []chunk {
	if numChunks <= 0 {
		numChunks = 1
	}
	orderGroups := make(map[int64][]models.OrderEvent)
	var orderKeys []int64
	for _, e := range events {
		if _, exists := orderGroups[e.OrderID]; !exists {
			orderKeys = append(orderKeys, e.OrderID)
		}
		orderGroups[e.OrderID] = append(orderGroups[e.OrderID], e)
	}

	if len(orderKeys) < numChunks {
		numChunks = len(orderKeys)
	}

	chunks := make([]chunk, numChunks)
	for i, key := range orderKeys {
		idx := i % numChunks
		chunks[idx].events = append(chunks[idx].events, orderGroups[key]...)
	}

	return chunks
}

func validateBasic(req dto.ImportOrderEventRequest) string {
	if req.OrderID <= 0 {
		return "order_id must be greater than 0"
	}
	if !models.IsValidStatus(models.OrderStatus(req.Status)) {
		return "unknown status value"
	}
	if req.EventAt.IsZero() {
		return "event_at is required"
	}
	return ""
}
