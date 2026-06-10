package ai

import (
	"context"
	"main/internal/dto"
)

type AIAdapter interface {
	AnalyzeException(ctx context.Context, input dto.ExceptionInput) (dto.ExceptionOutput, error)
	SummarizeReport(ctx context.Context, input dto.ExceptionOutput) (dto.ReportSummaryOutput, error)
	Ping(ctx context.Context) error
}
