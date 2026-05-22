package services

import (
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"sync"
)

type OrderEventService interface {
	ImportOrderEvents(reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error)
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

type workerResult struct {
	req    dto.ImportOrderEventRequest
	detail repositories.ProcessResultDetail
	err    error
}

func (s *orderEventService) ImportOrderEvents(reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error) {
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

	jobs := make(chan []dto.ImportOrderEventRequest, len(orderGroups))
	results := make(chan workerResult, len(validReqs))

	var wg sync.WaitGroup
	numWorkers := s.maxWorkers
	if len(orderGroups) < numWorkers {
		numWorkers = len(orderGroups)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for group := range jobs {
				for _, req := range group {
					event := models.OrderEvent{
						OrderID:   req.OrderID,
						NewStatus: models.OrderStatus(req.Status),
						UpdatedBy: req.UpdatedBy,
						EventAt:   req.EventAt,
					}

					detail, err := s.orderEventRepo.ProcessSingleEventTx(event)
					results <- workerResult{
						req:    req,
						detail: detail,
						err:    err,
					}
				}
			}
		}()
	}

	for _, events := range orderGroups {
		jobs <- events
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var processingErr error
	for wr := range results {
		if wr.err != nil {
			processingErr = wr.err
			resp.Rejected++
			resp.Errors = append(resp.Errors, dto.EventError{
				OrderID: wr.req.OrderID,
				Status:  wr.req.Status,
				Reason:  wr.err.Error(),
			})
			continue
		}

		switch wr.detail.Result {
		case repositories.Accepted:
			resp.Accepted++
		case repositories.Rejected:
			resp.Rejected++
			resp.Errors = append(resp.Errors, dto.EventError{
				OrderID: wr.req.OrderID,
				Status:  wr.req.Status,
				Reason:  wr.detail.Reason,
			})
		case repositories.Duplicate:
			resp.Duplicate++
			resp.Errors = append(resp.Errors, dto.EventError{
				OrderID: wr.req.OrderID,
				Status:  wr.req.Status,
				Reason:  wr.detail.Reason,
			})
		}
	}

	return resp, processingErr
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
