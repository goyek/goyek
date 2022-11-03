package goyek

import "strconv"

// Status of a task run.
type Status uint8

// Statuses of task run.
const (
	StatusNotRun Status = iota
	StatusPassed
	StatusFailed
	StatusSkipped
)

func (s Status) String() string {
	switch s {
	case StatusNotRun:
		return "NOOP"
	case StatusPassed:
		return "PASS"
	case StatusFailed:
		return "FAIL"
	case StatusSkipped:
		return "SKIP"
	}
	return "goyek.Status(" + strconv.Itoa(int(s)) + ")"
}
