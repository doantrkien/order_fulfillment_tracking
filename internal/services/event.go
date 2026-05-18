package services

import (
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"sync"
)

const maxWorkers = 7

type OrderEventService interface {
	ImportOrderEvents(reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error)
}

type orderEventService struct {
	orderEventRepo repositories.OrderEventRepository
}

func NewOrderEventService(orderEventRepo repositories.OrderEventRepository) OrderEventService {
	return &orderEventService{
		orderEventRepo: orderEventRepo,
	}
}

// workerResult pairs the original request with the repo outcome.
type workerResult struct {
	req    dto.ImportOrderEventRequest
	detail repositories.ProcessResultDetail
	err    error
}

func (s *orderEventService) ImportOrderEvents(reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error) {
	resp := dto.ImportOrderEventsResponse{
		Errors: []dto.EventError{},
	}

	// Separate valid and invalid events via basic validation
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

	// Worker pool
	jobs := make(chan workerResult, len(validReqs))
	results := make(chan workerResult, len(validReqs))

	// Start workers
	var wg sync.WaitGroup
	numWorkers := maxWorkers
	if len(validReqs) < numWorkers {
		numWorkers = len(validReqs)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Map DTO → model
				event := models.OrderEvent{
					OrderID:   job.req.OrderID,
					NewStatus: models.OrderStatus(job.req.Status),
					UpdatedBy: job.req.UpdatedBy,
					EventAt:   job.req.EventAt,
					// PreviousStatus is left zero — repo fetches from DB
				}

				detail, err := s.orderEventRepo.ProcessSingleEventTx(event)
				results <- workerResult{
					req:    job.req,
					detail: detail,
					err:    err,
				}
			}
		}()
	}

	// Send jobs
	for _, req := range validReqs {
		jobs <- workerResult{req: req}
	}
	close(jobs)

	// Wait for workers to finish, then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate results
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

// validateBasic performs service-layer input validation.
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
