package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"main/internal/ai"
	dto_ai "main/internal/dto/ai"
	"main/internal/models"
	"main/internal/repositories"

	"gorm.io/datatypes"
)

type EvaluationWorker struct {
	analyzer   *ai.ExceptionAnalyzer
	evalRepo   repositories.AIEvaluationRepository
	maxWorkers int
}

func NewEvaluationWorker(analyzer *ai.ExceptionAnalyzer, evalRepo repositories.AIEvaluationRepository, maxWorkers int) *EvaluationWorker {
	if maxWorkers <= 0 {
		maxWorkers = 3
	}
	return &EvaluationWorker{
		analyzer:   analyzer,
		evalRepo:   evalRepo,
		maxWorkers: maxWorkers,
	}
}

func (w *EvaluationWorker) Run(runID int64, cases []dto_ai.EvaluationCase) {
	ctx := context.Background()

	runRecord, err := w.evalRepo.GetRunByID(ctx, runID)
	if err != nil {
		log.Printf("[EvaluationWorker] Failed to get run record %d: %v", runID, err)
		return
	}

	runRecord.Status = models.EVAL_STATUS_IN_PROGRESS
	runRecord.UpdatedAt = time.Now()
	if err := w.evalRepo.UpdateRun(ctx, runRecord); err != nil {
		log.Printf("[EvaluationWorker] Failed to update run status to IN_PROGRESS: %v", err)
		return
	}

	numCases := len(cases)
	jobs := make(chan dto_ai.EvaluationCase, numCases)
	results := make(chan *ai.CaseResult, numCases)

	var wg sync.WaitGroup
	var hasFatalPanic bool
	var mu sync.Mutex

	for i := 0; i < w.maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					log.Printf("[EvaluationWorker %d] Panic recovered: %v", workerID, r)
					mu.Lock()
					hasFatalPanic = true
					mu.Unlock()
				}
			}()

			for evalCase := range jobs {
				startTime := time.Now()

				aiCtx := buildAIContextFromSyntheticInput(evalCase.SyntheticInput)

				analysisResult, err := w.analyzer.Analyze(ctx, aiCtx)

				fmt.Printf("Analysis Result: %+v\n", analysisResult)

				var caseResult *ai.CaseResult
				var errMsg *string

				if err != nil {
					errMsgStr := err.Error()
					errMsg = &errMsgStr
					caseResult = &ai.CaseResult{
						CaseID:           evalCase.CaseID,
						IsPassed:         false,
						ExpectedType:     evalCase.ExpectedExceptionType,
						ExpectedSeverity: evalCase.ExpectedSeverity,
						LatencyMs:        int(time.Since(startTime).Milliseconds()),
						ErrorMessage:     errMsgStr,
					}
				} else {
					caseResult = ai.CompareResult(evalCase, analysisResult)
					if caseResult.ErrorMessage != "" {
						errMsgStr := caseResult.ErrorMessage
						errMsg = &errMsgStr
					}
				}

				detailStatus := models.EVAL_DETAIL_FAILED
				if caseResult.IsPassed {
					detailStatus = models.EVAL_DETAIL_PASSED
				}

				var actualOutputJSON []byte
				if analysisResult != nil {
					actualOutputJSON = []byte(fmt.Sprintf(`{"exception_type":"%s", "severity":"%s", "fallback_used":%v}`,
						analysisResult.ExceptionType, analysisResult.Severity, analysisResult.FallbackUsed))
				} else {
					actualOutputJSON = []byte(`{}`)
				}

				inputJSON := []byte(fmt.Sprintf(`{"order_id":%d}`, evalCase.SyntheticInput.OrderID)) // minimal input representation
				expectedJSON := []byte(fmt.Sprintf(`{"expected_exception_type":"%s", "expected_severity":"%s"}`, evalCase.ExpectedExceptionType, evalCase.ExpectedSeverity))

				detail := &models.AIEvaluationResultDetail{
					RunID:          runID,
					Input:          datatypes.JSON(inputJSON),
					ExpectedOutput: datatypes.JSON(expectedJSON),
					ActualOutput:   datatypes.JSON(actualOutputJSON),
					Status:         detailStatus,
					LatencyMs:      caseResult.LatencyMs,
					ErrorMessage:   errMsg,
				}

				if err := w.evalRepo.SaveDetail(ctx, detail); err != nil {
					log.Printf("[EvaluationWorker %d] Failed to save detail for case %s: %v", workerID, evalCase.CaseID, err)
				}

				results <- caseResult
			}
		}(i)
	}

	for _, c := range cases {
		jobs <- c
	}
	close(jobs)

	wg.Wait()
	close(results)

	var allResults []*ai.CaseResult
	for r := range results {
		allResults = append(allResults, r)
	}

	// 6. Calculate Metrics and Update Final Status
	mu.Lock()
	panicOccurred := hasFatalPanic
	mu.Unlock()

	if panicOccurred {
		runRecord.Status = models.EVAL_STATUS_FAILED
	} else {
		runRecord.Status = models.EVAL_STATUS_COMPLETED
	}

	metrics := ai.CalculateRunMetrics(allResults)

	runRecord.PassedCases = metrics.PassedCases
	runRecord.FailedCases = metrics.FailedCases
	runRecord.FallbackCount = metrics.FallbackCount
	runRecord.AccuracyRate = metrics.AccuracyRate
	runRecord.AvgLatencyMs = metrics.AvgLatencyMs
	runRecord.UpdatedAt = time.Now()

	if err := w.evalRepo.UpdateRun(ctx, runRecord); err != nil {
		log.Printf("[EvaluationWorker] Failed to update final run metrics: %v", err)
	} else {
		log.Printf("[EvaluationWorker] Run %d finished. Status: %s. Accuracy: %.2f%%", runID, runRecord.Status, runRecord.AccuracyRate)
	}
}

func buildAIContextFromSyntheticInput(input dto_ai.ExceptionPromptContext) *models.AIContext {
	aiCtx := &models.AIContext{
		OrderID:         input.OrderID,
		CurrentStatus:   models.OrderStatus(input.CurrentStatus),
		TotalAmount:     input.TotalAmount,
		CustomerName:    input.CustomerName,
		ShippingAddress: input.ShippingAddress,
	}

	if t, err := time.Parse(time.RFC3339, input.CreatedAt); err == nil {
		aiCtx.CreatedAt = t
	} else {
		aiCtx.CreatedAt = time.Now().Add(-time.Hour * 24 * 7)
	}

	var events []models.AIEvent
	for i, eh := range input.EventTimeline {
		var prevStatus models.OrderStatus
		if eh.FromStatus != "" {
			prevStatus = models.OrderStatus(eh.FromStatus)
		}

		eventTime := time.Now()
		if t, err := time.Parse(time.RFC3339, eh.EventAt); err == nil {
			eventTime = t
		} else {
			eventTime = aiCtx.CreatedAt.Add(time.Duration(i) * time.Hour)
		}

		events = append(events, models.AIEvent{
			EventAt:        eventTime,
			PreviousStatus: prevStatus,
			NewStatus:      models.OrderStatus(eh.ToStatus),
			UpdatedBy:      eh.UpdatedBy,
		})
	}

	if note := strings.TrimSpace(input.DriverNotes); note != "" && len(events) > 0 {
		events[len(events)-1].DriverNote = &note
	}

	aiCtx.Events = events

	return aiCtx
}
