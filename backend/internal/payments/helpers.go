package payments

import (
	"strconv"
	"time"
)

// itoa converts an int64 to a decimal string.
func itoa(n int64) string { return strconv.FormatInt(n, 10) }

// reqID builds a short unique suffix for a user/charge id.
func reqID(userID string) string {
	suffix := userID
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	return suffix + itoa(time.Now().UnixNano()%10000)
}
