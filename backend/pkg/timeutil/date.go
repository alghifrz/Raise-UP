package timeutil

import (
	"fmt"
	"time"
)

// Jakarta is the application calendar timezone for date-only filters.
// Date-only query params (YYYY-MM-DD) are interpreted as Asia/Jakarta calendar days,
// then converted to timestamptz bounds for filtering.
var Jakarta *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		Jakarta = time.FixedZone("WIB", 7*60*60)
		return
	}
	Jakarta = loc
}

// ParseDateOnly parses YYYY-MM-DD as a Jakarta calendar date (midnight local).
func ParseDateOnly(value string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", value, Jakarta)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date, expected YYYY-MM-DD")
	}
	return t, nil
}

// DayRange converts optional from/to date strings into inclusive/exclusive timestamptz bounds.
// from: timestamp >= start of from-day (Jakarta)
// to:   timestamp <  start of day after to-day (Jakarta), making to inclusive for the full day
func DayRange(from, to string) (fromAt *time.Time, toExclusive *time.Time, err error) {
	var fromTime, toTime time.Time
	hasFrom := from != ""
	hasTo := to != ""

	if hasFrom {
		fromTime, err = ParseDateOnly(from)
		if err != nil {
			return nil, nil, err
		}
	}
	if hasTo {
		toTime, err = ParseDateOnly(to)
		if err != nil {
			return nil, nil, err
		}
	}
	if hasFrom && hasTo && fromTime.After(toTime) {
		return nil, nil, fmt.Errorf("from must be less than or equal to to")
	}

	if hasFrom {
		fromAt = &fromTime
	}
	if hasTo {
		next := toTime.AddDate(0, 0, 1)
		toExclusive = &next
	}
	return fromAt, toExclusive, nil
}
