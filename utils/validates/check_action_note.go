package validates

import (
	"strings"
	"unicode"
)

func IsNoteActionable(note string) bool {
	s := strings.TrimSpace(note)

	if len([]rune(s)) < 10 {
		return false
	}

	letterCount := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letterCount++
		}
	}
	if float64(letterCount)/float64(len([]rune(s))) < 0.4 {
		return false
	}

	if len(strings.Fields(s)) < 3 {
		return false
	}

	return true
}
