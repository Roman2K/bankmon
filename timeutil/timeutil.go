package timeutil

import "time"

func BeginningOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	loc := t.Location()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func EndOfPrevDay(t time.Time) time.Time {
	return BeginningOfDay(t).Add(-1 * time.Nanosecond)
}
