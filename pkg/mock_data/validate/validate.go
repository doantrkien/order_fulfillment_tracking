package validate

import "sync"

var allowed = map[string][]string{
	"created": {"paid", "cancelled"},
	"paid":    {"packed", "refunded"},
	"packed":  {"shipped"},
	"shipped": {"delivered"},
}

func IsValidStatus(before, after string) bool {
	for _, v := range allowed[before] {
		if v == after {
			return true
		}
	}

	return false
}

func IsDuplicate(seen map[int64]bool, id int64, mu *sync.Mutex) bool {
	mu.Lock()
	defer mu.Unlock()

	if seen[id] {
		return true
	}

	seen[id] = true
	return false
}
