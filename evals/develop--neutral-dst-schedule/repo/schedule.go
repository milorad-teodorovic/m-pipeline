package schedule

import "time"

// Daily is a job that runs every day at Hour:Minute UTC.
type Daily struct {
	Hour   int
	Minute int
}

// NextRun returns the first run of d strictly after after.
func (d Daily) NextRun(after time.Time) time.Time {
	after = after.UTC()
	t := time.Date(after.Year(), after.Month(), after.Day(), d.Hour, d.Minute, 0, 0, time.UTC)
	if !t.After(after) {
		t = t.AddDate(0, 0, 1)
	}
	return t
}
