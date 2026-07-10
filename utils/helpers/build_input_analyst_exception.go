package helpers

import (
	dto "main/internal/dto/ai"
	"main/internal/models"
	"strings"
	"time"
)

// func buildExceptionInput(aiCtx *models.AIContext, notes string) dto.ExceptionInput {
func BuildExceptionInput(aiCtx *models.AIContext) dto.ExceptionPromptContext {
	eventHistory := make([]dto.EventTimelineEntry, 0, len(aiCtx.Events))
	for _, e := range aiCtx.Events {
		eventHistory = append(eventHistory, dto.EventTimelineEntry{
			FromStatus: string(e.PreviousStatus),
			ToStatus:   string(e.NewStatus),
			EventAt:    e.EventAt.Format(time.RFC3339),
			UpdatedBy:  e.UpdatedBy,
		})
	}

	// allNotes := []string{}
	// // if strings.TrimSpace(notes) != "" {
	// // 	allNotes = append(allNotes, strings.TrimSpace(notes))
	// // }
	// for _, e := range aiCtx.Events {
	// 	if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
	// 		allNotes = append(allNotes, strings.TrimSpace(*e.DriverNote))
	// 	}
	// }
	// errorMessage := strings.Join(allNotes, "; ")

	var driverNotes string
	for i := len(aiCtx.Events) - 1; i >= 0; i-- {
		e := aiCtx.Events[i]
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			driverNotes = strings.TrimSpace(*e.DriverNote)
			break
		}
	}

	return dto.ExceptionPromptContext{
		OrderID:         aiCtx.OrderID,
		CurrentStatus:   string(aiCtx.CurrentStatus),
		TotalAmount:     aiCtx.TotalAmount,
		CustomerName:    aiCtx.CustomerName,
		ShippingAddress: aiCtx.ShippingAddress,
		CreatedAt:       aiCtx.CreatedAt.Format(time.RFC3339),
		DriverNotes:     driverNotes,
		EventTimeline:   eventHistory,
	}
}
