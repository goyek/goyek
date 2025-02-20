package goyek

import (
	"context"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// A is a type passed to [Task.Action] functions to manage task state
// and support formatted task logs.
//
// A task ends when its action function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf.
// Those methods must be called only from the goroutine running the action function.
//
// The other reporting methods, such as the variations of Log and Error,
// may be called simultaneously from multiple goroutines.
type A struct {
	ctx         context.Context
	name        string
	output      io.Writer
	logger      Logger
	mu          sync.Mutex
	failed      bool
	failedCall  func()
	skipped     bool
	skippedCall func()
	cleanups    []func()
}

// Context returns the run context.
func (a *A) Context() context.Context {
	return a.ctx
}

// WithContext returns a shallow copy of a with its context changed
// to ctx. The provided ctx must be non-nil.
func (a *A) WithContext(ctx context.Context) *A {
	if ctx == nil {
		panic("nil context")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	result := &A{
		ctx:         ctx,
		name:        a.name,
		output:      a.output,
		logger:      a.logger,
		failed:      a.failed,
		failedCall:  a.Fail,
		skipped:     a.skipped,
		skippedCall: a.SkipNow,
	}

	a.cleanups = append(a.cleanups, func() {
		result.callCleanups()
	})

	return result
}

// Name returns the name of the running task.
func (a *A) Name() string {
	return a.name
}

// Output returns the destination used for printing messages.
func (a *A) Output() io.Writer {
	return a.output
}

// Log formats its arguments using default formatting, analogous to Println,
// and writes the text to [A.Output]. A final newline is added.
func (a *A) Log(args ...interface{}) {
	a.logger.Log(a.output, args...)
}

// Logf formats its arguments according to the format, analogous to Printf,
// and writes the text to [A.Output]. A final newline is added.
func (a *A) Logf(format string, args ...interface{}) {
	a.logger.Logf(a.output, format, args...)
}

// Error is equivalent to [A.Log] followed by [A.Fail].
func (a *A) Error(args ...interface{}) {
	if l, ok := a.logger.(interface {
		Error(w io.Writer, args ...interface{})
	}); ok {
		l.Error(a.output, args...)
	} else {
		a.logger.Log(a.output, args...)
	}

	a.Fail()
}

// Errorf is equivalent to [A.Logf] followed by [A.Fail].
func (a *A) Errorf(format string, args ...interface{}) {
	if l, ok := a.logger.(interface {
		Errorf(w io.Writer, format string, args ...interface{})
	}); ok {
		l.Errorf(a.output, format, args...)
	} else {
		a.logger.Logf(a.output, format, args...)
	}

	a.Fail()
}

// Failed reports whether the function has failed.
func (a *A) Failed() bool {
	a.mu.Lock()
	res := a.failed
	a.mu.Unlock()
	return res
}

// Fail marks the function as having failed but continues execution.
func (a *A) Fail() {
	var call func()
	a.mu.Lock()
	a.failed = true
	call = a.failedCall
	a.mu.Unlock()
	if call != nil {
		call()
	}
}

// Fatal is equivalent to [A.Log] followed by [A.FailNow].
func (a *A) Fatal(args ...interface{}) {
	if l, ok := a.logger.(interface {
		Fatal(w io.Writer, args ...interface{})
	}); ok {
		l.Fatal(a.output, args...)
	} else {
		a.logger.Log(a.output, args...)
	}

	a.FailNow()
}

// Fatalf is equivalent to [A.Logf] followed by [A.FailNow].
func (a *A) Fatalf(format string, args ...interface{}) {
	if l, ok := a.logger.(interface {
		Fatalf(w io.Writer, format string, args ...interface{})
	}); ok {
		l.Fatalf(a.output, format, args...)
	} else {
		a.logger.Logf(a.output, format, args...)
	}

	a.FailNow()
}

// FailNow marks the function as having failed
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// It finishes the whole flow execution.
// FailNow must be called from the goroutine running the [Task.Action] function,
// not from other goroutines created during its execution.
// Calling FailNow does not stop those other goroutines.
func (a *A) FailNow() {
	a.Fail()
	runtime.Goexit()
}

// Skipped reports whether the task was skipped.
func (a *A) Skipped() bool {
	a.mu.Lock()
	res := a.skipped
	a.mu.Unlock()
	return res
}

// Skip is equivalent to [A.Log] followed by [A.SkipNow].
func (a *A) Skip(args ...interface{}) {
	if l, ok := a.logger.(interface {
		Skip(w io.Writer, args ...interface{})
	}); ok {
		l.Skip(a.output, args...)
	} else {
		a.logger.Log(a.output, args...)
	}

	a.SkipNow()
}

// Skipf is equivalent to [A.Logf] followed by [A.SkipNow].
func (a *A) Skipf(format string, args ...interface{}) {
	if l, ok := a.logger.(interface {
		Skipf(w io.Writer, format string, args ...interface{})
	}); ok {
		l.Skipf(a.output, format, args...)
	} else {
		a.logger.Logf(a.output, format, args...)
	}
	a.SkipNow()
}

// SkipNow marks the task as having been skipped
// and stops its execution by calling runtime.Goexit
// (which then runs all deferred calls in the current goroutine).
// If a test fails (see Error, Errorf, Fail) and is then skipped,
// it is still considered to have failed.
// The flow execution will continue at the next task.
// See also [A.FailNow].
// SkipNow must be called from the goroutine running the [Task.Action] function,
// not from other goroutines created during its execution.
// Calling SkipNow does not stop those other goroutines.
func (a *A) SkipNow() {
	var call func()
	a.mu.Lock()
	a.skipped = true
	call = a.skippedCall
	a.mu.Unlock()
	if call != nil {
		call()
	}
	runtime.Goexit()
}

// Helper marks the calling function as a helper function.
// It calls logger's Helper method if implemented.
// By default, when printing file and line information, that function will be skipped.
func (a *A) Helper() {
	if h, ok := a.logger.(interface {
		Helper()
	}); ok {
		h.Helper()
	}
}

// Cleanup registers a function to be called when [Task.Action] function completes.
// Cleanup functions will be called in the last-added first-called order.
func (a *A) Cleanup(fn func()) {
	a.mu.Lock()
	a.cleanups = append(a.cleanups, fn)
	a.mu.Unlock()
}

// Setenv calls os.Setenv(key, value) and uses Cleanup to restore the environment variable
// to its original value after the action.
func (a *A) Setenv(key, value string) {
	a.Helper()
	prevValue, ok := os.LookupEnv(key)

	if err := os.Setenv(key, value); err != nil {
		a.Fatalf("cannot set environment variable: %v", err)
	}

	if ok {
		a.Cleanup(func() {
			os.Setenv(key, prevValue)
		})
	} else {
		a.Cleanup(func() {
			os.Unsetenv(key)
		})
	}
}

// TempDir returns a temporary directory for the action to use.
// The directory is automatically removed by Cleanup when the action completes.
// Each subsequent call to TempDir returns a unique directory;
// if the directory creation fails, TempDir terminates the action by calling Fatal.
func (a *A) TempDir() string {
	a.Helper()
	// Drop unusual characters (such as path separators or
	// characters interacting with globs) from the directory name to
	// avoid surprising os.MkdirTemp behavior.
	mapper := func(r rune) rune {
		if r < utf8.RuneSelf {
			const allowed = "!#$%&()+,-.=@^_{}~ "
			if '0' <= r && r <= '9' ||
				'a' <= r && r <= 'z' ||
				'A' <= r && r <= 'Z' {
				return r
			}
			if strings.ContainsRune(allowed, r) {
				return r
			}
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return -1
	}
	name := strings.Map(mapper, a.Name())

	dir, err := os.MkdirTemp("", "goyek-"+name+"-*")
	if err != nil {
		a.Fatalf("cannot create temporary directory: %v", err)
	}
	a.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			a.Errorf("TempDir RemoveAll cleanup: %v", err)
		}
	})
	return dir
}

func (a *A) run(action func(a *A)) (finished bool, panicVal interface{}, panicStack []byte) {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		defer a.runCleanups(&finished, &panicVal, &panicStack)
		defer func() {
			if finished {
				return
			}
			panicVal = recover()
			panicStack = debug.Stack()
		}()
		action(a)
		finished = true
	}()
	<-ch
	return
}

func (a *A) runCleanups(finished *bool, panicVal *interface{}, panicStack *[]byte) {
	// we capture only the first panic
	cleanupFinished := false
	if *finished {
		defer func() {
			if cleanupFinished {
				return
			}
			*panicVal = recover()
			*panicStack = debug.Stack()
			*finished = false
		}()
	} else {
		defer func() {
			_ = recover() // ignore next panics
		}()
	}

	// Make sure that if a cleanup function panics,
	// we still run the remaining cleanup functions.
	defer func() {
		a.mu.Lock()
		recur := len(a.cleanups) > 0
		a.mu.Unlock()
		if recur {
			a.runCleanups(finished, panicVal, panicStack)
		}
	}()

	a.callCleanups()
	cleanupFinished = true
}

func (a *A) callLastCleanup() bool {
	var cleanup func()
	a.mu.Lock()
	if len(a.cleanups) > 0 {
		last := len(a.cleanups) - 1
		cleanup = a.cleanups[last]
		a.cleanups = a.cleanups[:last]
	}
	a.mu.Unlock()
	if cleanup == nil {
		return false
	}
	cleanup()
	return true
}

func (a *A) callCleanups() {
	for a.callLastCleanup() {
	}
}
