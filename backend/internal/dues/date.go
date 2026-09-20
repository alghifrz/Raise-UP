package dues

import (
	"time"

	"github.com/diuk/raiseup/pkg/timeutil"
)

// Jakarta re-exports the shared Asia/Jakarta location used by paid_at date filters.
var Jakarta = timeutil.Jakarta

// ParseDateOnly parses YYYY-MM-DD as a Jakarta calendar date.
func ParseDateOnly(value string) (time.Time, error) {
	return timeutil.ParseDateOnly(value)
}

// DayRange converts optional from/to date strings into inclusive/exclusive timestamptz bounds
// for paid_at filtering (same Jakarta calendar-day semantics as finance).
func DayRange(from, to string) (fromAt *time.Time, toExclusive *time.Time, err error) {
	return timeutil.DayRange(from, to)
}
