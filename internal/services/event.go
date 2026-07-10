package services

import (
	"context"
	dto_api "main/internal/dto/api"
	"main/internal/models"
	"main/internal/repositories"
	"sort"
	"sync"
	"time"
)

type OrderEventService interface {
	ImportOrderEvents(ctx context.Context, reqs []dto_api.ImportOrderEventRequest) (dto_api.ImportOrderEventsResponse, error)
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

func (s *orderEventService) ImportOrderEvents(ctx context.Context, reqs []dto_api.ImportOrderEventRequest) (dto_api.ImportOrderEventsResponse, error) {
	resp := dto_api.ImportOrderEventsResponse{
		Errors: []dto_api.EventError{},
	}

	var validReqs []dto_api.ImportOrderEventRequest
	for _, req := range reqs {
		if reason := validateBasic(req); reason != "" {
			resp.Rejected++
			resp.Errors = append(resp.Errors, dto_api.EventError{
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

	orderGroups := make(map[int64][]dto_api.ImportOrderEventRequest)
	for _, req := range validReqs {
		orderGroups[req.OrderID] = append(orderGroups[req.OrderID], req)
	}
	for _, events := range orderGroups {
		sort.Slice(events, func(i, j int) bool {
			return events[i].EventAt.Before(events[j].EventAt)
		})
	}

	batches := splitIntoBatches(orderGroups, s.maxWorkers)
	if len(batches) == 0 {
		return resp, nil
	}

	jobs := make(chan []models.OrderEvent, len(batches))
	resultsChan := make(chan batchResult, len(batches))

	var wg sync.WaitGroup
	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for events := range jobs {
				select {
				case <-ctx.Done():
					resultsChan <- batchResult{err: ctx.Err()}
					continue
				default:
				}
				details, err := s.orderEventRepo.ProcessBatchEventsTx(ctx, events)
				resultsChan <- batchResult{details: details, err: err}
			}
		}()
	}

	for _, batch := range batches {
		jobs <- batch
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

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
				resp.Errors = append(resp.Errors, dto_api.EventError{
					OrderID: detail.OrderID,
					Status:  detail.Status,
					Reason:  detail.Reason,
				})
			case repositories.Duplicate:
				resp.Duplicate++
				resp.Errors = append(resp.Errors, dto_api.EventError{
					OrderID: detail.OrderID,
					Status:  detail.Status,
					Reason:  detail.Reason,
				})
			}
		}
	}

	return resp, processingErr
}

func splitIntoBatches(orderGroups map[int64][]dto_api.ImportOrderEventRequest, numBatches int) [][]models.OrderEvent {
	if numBatches <= 0 {
		numBatches = 1
	}
	if len(orderGroups) < numBatches {
		numBatches = len(orderGroups)
	}

	batches := make([][]models.OrderEvent, numBatches)

	var orderIDs []int64
	for id := range orderGroups {
		orderIDs = append(orderIDs, id)
	}
	sort.Slice(orderIDs, func(i, j int) bool {
		return orderIDs[i] < orderIDs[j]
	})

	i := 0
	for _, id := range orderIDs {
		reqs := orderGroups[id]
		idx := i % numBatches
		for _, req := range reqs {
			// If the client sends event_at without a timezone offset (e.g. "2026-05-31T14:55:00"),
			// Go's JSON decoder parses it as UTC. We re-interpret it as Asia/Ho_Chi_Minh local
			// time and convert to UTC so the value stored in PostgreSQL is correct.
			eventAt := req.EventAt
			if eventAt.Location() == time.UTC {
				// Treat the wall-clock value as HCM local time, then shift to UTC.
				eventAt = time.Date(
					eventAt.Year(), eventAt.Month(), eventAt.Day(),
					eventAt.Hour(), eventAt.Minute(), eventAt.Second(), eventAt.Nanosecond(),
					loc,
				).UTC()
			}
			var driverIDPtr *int64
			if req.DriverID != 0 {
				driverIDPtr = &req.DriverID
			}
			var driverNotePtr *string
			if req.DriverNote != "" {
				driverNotePtr = &req.DriverNote
			}

			batches[idx] = append(batches[idx], models.OrderEvent{
				OrderID:    req.OrderID,
				NewStatus:  models.OrderStatus(req.Status),
				UpdatedBy:  req.UpdatedBy,
				EventAt:    eventAt,
				DriverID:   driverIDPtr,
				DriverNote: driverNotePtr,
			})
		}
		i++
	}

	return batches
}

func validateBasic(req dto_api.ImportOrderEventRequest) string {
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
