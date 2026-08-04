package id

import (
	"fmt"
	"time"
)

// New returns an id like "<prefix>_<unixms>_<nanosuffix>" that is unique
// per call within a single process.
func New(prefix string) string {
	now := time.Now()
	return fmt.Sprintf("%s_%d_%d", prefix, now.UnixMilli(), now.Nanosecond())
}
