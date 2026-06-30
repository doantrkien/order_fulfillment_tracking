package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"main/internal/ai"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"

	"gorm.io/datatypes"
)

// EvaluationWorker manages a pool of background workers to process an AI evaluation batch.
type EvaluationWorker struct {
	analyzer   *ai.ExceptionAnalyzer
	evalRepo   repositories.AIEvaluationRepository
	maxWorkers int
}

func NewEvaluationWorker(analyzer *ai.ExceptionAnalyzer, evalRepo repositories.AIEvaluationRepository, maxWorkers int) *EvaluationWorker {
	if maxWorkers <= 0 {
		maxWorkers = 3 // default to 3 workers if invalid
	}
	return &EvaluationWorker{
		analyzer:   analyzer,
		evalRepo:   evalRepo,
		maxWorkers: maxWorkers,
	}
}

// Run executes the evaluation batch asynchronously.
// It fetches the run from DB, updates status to IN_PROGRESS,
// spins up worker goroutines to process cases, and finally aggregates and saves metrics.
func (w *EvaluationWorker) Run(runID int64, cases []dto.EvaluationCase) {
	ctx := context.Background()

	// 1. Fetch Run Record
	runRecord, err := w.evalRepo.GetRunByID(ctx, runID)
	if err != nil {
		log.Printf("[EvaluationWorker] Failed to get run record %d: %v", runID, err)
		return
	}

	// 2. Mark as IN_PROGRESS
	runRecord.Status = models.EVAL_STATUS_IN_PROGRESS
	runRecord.UpdatedAt = time.Now()
	if err := w.evalRepo.UpdateRun(ctx, runRecord); err != nil {
		log.Printf("[EvaluationWorker] Failed to update run status to IN_PROGRESS: %v", err)
		return
	}

	// 3. Setup concurrency primitives
	numCases := len(cases)
	jobs := make(chan dto.EvaluationCase, numCases)
	results := make(chan *ai.CaseResult, numCases)

	var wg sync.WaitGroup
	var hasFatalPanic bool
	var mu sync.Mutex // To safely update hasFatalPanic if needed

	// 4. Spawn Workers
	for i := 0; i < w.maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Panic recovery for each worker
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[EvaluationWorker %d] Panic recovered: %v", workerID, r)
					mu.Lock()
					hasFatalPanic = true
					mu.Unlock()
				}
			}()

			for evalCase := range jobs {
				// fmt.Println("Processing case:", evalCase)
				startTime := time.Now()

				// 4.1 Build AI Context from synthetic input
				aiCtx := buildAIContextFromSyntheticInput(evalCase.SyntheticInput)

				// fmt.Printf("Processing case %s\n", evalCase.CaseID)
				// jsonBytes, _ := json.MarshalIndent(aiCtx, "", "  ")
				// fmt.Printf("AI Context: %s\n", string(jsonBytes))

				// 4.2 Run AI Analysis
				// analysisResult, err := w.analyzer.Analyze(ctx, aiCtx, "")
				analysisResult, err := w.analyzer.Analyze(ctx, aiCtx)

				fmt.Printf("Analysis Result: %+v\n", analysisResult)

				// 4.3 Compare Result
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

				// 4.4 Determine detail status
				detailStatus := models.EVAL_DETAIL_FAILED
				if caseResult.IsPassed {
					detailStatus = models.EVAL_DETAIL_PASSED
				}

				// 4.5 Build Actual Output JSON
				var actualOutputJSON []byte
				if analysisResult != nil {
					// We only need basic fields for actual_output logging, or we can just martial the AnalysisResult
					actualOutputJSON = []byte(fmt.Sprintf(`{"exception_type":"%s", "severity":"%s", "fallback_used":%v}`,
						analysisResult.ExceptionType, analysisResult.Severity, analysisResult.FallbackUsed))
				} else {
					actualOutputJSON = []byte(`{}`)
				}

				// 4.6 Save Detail to DB
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

				// 4.7 Push to results
				results <- caseResult
			}
		}(i)
	}

	// 5. Enqueue Jobs and Close Jobs Channel
	for _, c := range cases {
		jobs <- c
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Collect results
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

// buildAIContextFromSyntheticInput maps the synthetic exception input into models.AIContext
func buildAIContextFromSyntheticInput(input dto.ExceptionInput) *models.AIContext {
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
	for i, eh := range input.EventHistory {
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

	// Map top-level driver_notes into the last event's DriverNote so that
	// hasAnyDriverNote() can detect it and allow AI analysis to run.
	if note := strings.TrimSpace(input.DriverNotes); note != "" && len(events) > 0 {
		events[len(events)-1].DriverNote = &note
	}

	aiCtx.Events = events

	return aiCtx
}
