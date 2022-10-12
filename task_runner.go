package goyek

import (
	"fmt"
	"io"
	"time"
)

// taskRunner is used to run a Action.
type taskRunner struct{}

// runResult contains the results of a Action run.
type runResult struct {
	Failed   bool
	Skipped  bool
	Duration time.Duration
}

// Run runs the action.
func (r taskRunner) Run(tf *TF, action func(tf *TF)) runResult {
	finished := make(chan runResult, 1)
	go func() {
		from := time.Now()
		defer func() {
			if r := recover(); r != nil {
				txt := fmt.Sprintf("panic: %v", r)
				const skipUntilPanic = 3
				txt = tf.decorate(txt, skipUntilPanic)
				io.WriteString(tf.writer, txt) //nolint // not checking errors when writing to output
				tf.failed = true
			}
			result := runResult{
				Failed:   tf.failed,
				Skipped:  tf.skipped,
				Duration: time.Since(from),
			}
			finished <- result
		}()
		action(tf)
	}()
	return <-finished
}
