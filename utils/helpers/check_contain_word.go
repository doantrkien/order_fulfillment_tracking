package helpers

import (
	"strings"
)

func ContainsAny(s string, keywords []string) bool {
	normalizedS := RemoveAccents(s)
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
		if strings.Contains(normalizedS, RemoveAccents(kw)) {
			return true
		}
	}
	return false
}
