package validates

import (
	"main/internal/models"
	"strings"
)

func HasAnyDriverNote(aiCtx *models.AIContext) bool {
	for _, e := range aiCtx.Events {
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			return true
		}
	}
	return false
}
