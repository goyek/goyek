package goyek_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/goyek/goyek/v3"
)

func TestA_Context_cancels_before_cleanup(t *testing.T) {
	sb := &strings.Builder{}

	res := goyek.NewRunner(func(a *goyek.A) {
		ctx := a.Context()

		a.Cleanup(func() {
			// The context should be canceled when cleanup functions are called.
			select {
			case <-ctx.Done():
				a.Log("context was canceled before cleanup")
			default:
				a.Log("context was NOT canceled before cleanup")
			}
		})

		// Context should not be canceled during action execution.
		select {
		case <-ctx.Done():
			t.Error("context should not be canceled during action execution")
		default:
			a.Log("context is active during action")
		}
	})(goyek.Input{Context: context.Background(), Output: goyek.SyncWriter(sb), Logger: goyek.FmtLogger{}})

	if res.Status != goyek.StatusPassed {
		t.Errorf("status was %s but want %s", res.Status, goyek.StatusPassed)
	}

	want := "context is active during action\ncontext was canceled before cleanup\n"
	if got := sb.String(); got != want {
		t.Errorf("output was %q but want %q", got, want)
	}
}

func TestA_WithContext(t *testing.T) {
	testCases := []struct {
		desc        string
		fn          func(a, a2 *goyek.A)
		wantStatus  goyek.Status
		wantFailed  bool
		wantSkipped bool
	}{
		{
			desc:       "Pass",
			fn:         func(_, _ *goyek.A) {},
			wantStatus: goyek.StatusPassed,
		},
		{
			desc: "Fail",
			fn: func(a, _ *goyek.A) {
				a.FailNow()
			},
			wantStatus: goyek.StatusFailed,
			wantFailed: true,
		},
		{
			desc: "FailDerived",
			fn: func(_, a2 *goyek.A) {
				a2.FailNow()
			},
			wantStatus: goyek.StatusFailed,
			wantFailed: true,
		},
		{
			desc: "Skip",
			fn: func(a, _ *goyek.A) {
				a.SkipNow()
			},
			wantStatus:  goyek.StatusSkipped,
			wantSkipped: true,
		},
		{
			desc: "SkipDerived",
			fn: func(_, a2 *goyek.A) {
				a2.SkipNow()
			},
			wantStatus:  goyek.StatusSkipped,
			wantSkipped: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			sb := &strings.Builder{}

			ctx := context.Background()
			type ctxKey struct{}
			newCtx := context.WithValue(ctx, ctxKey{}, 0)

			var (
				got, got2         context.Context
				failed, skipped   bool
				failed2, skipped2 bool
			)
			res := goyek.NewRunner(func(a *goyek.A) {
				a2 := a.WithContext(newCtx)
				got = a.Context()   // ctx
				got2 = a2.Context() // newCtx
				a2.Log("1")
				a.Cleanup(func() {
					skipped = a.Skipped()
					failed = a.Failed()
					a.Log("3")
				})
				a2.Cleanup(func() {
					skipped2 = a2.Skipped()
					failed2 = a2.Failed()
					a2.Log("2")
				})
				tc.fn(a, a2)
			})(goyek.Input{Context: ctx, Output: goyek.SyncWriter(sb), Logger: goyek.FmtLogger{}})

			if res.Status != tc.wantStatus {
				t.Errorf("status was %s but want %s", res.Status, tc.wantStatus)
			}
			// Check that context values are preserved rather than exact equality
			// since the contexts are now derived with cancellation capability.
			if got.Value(ctxKey{}) != ctx.Value(ctxKey{}) {
				t.Errorf("original Context value mismatch")
			}
			if got2.Value(ctxKey{}) != newCtx.Value(ctxKey{}) {
				t.Errorf("derived Context value mismatch")
			}
			if out, want := sb.String(), "1\n2\n3\n"; out != want {
				t.Errorf("logging or cleanup failed, out was %q but want %q", out, want)
			}
			if failed != tc.wantFailed {
				t.Errorf("original Failed returned %v but want %v", failed, tc.wantFailed)
			}
			if skipped != tc.wantSkipped {
				t.Errorf("original Skipped returned %v but want %v", skipped, tc.wantSkipped)
			}
			if failed2 != tc.wantFailed {
				t.Errorf("derived Failed returned %v but want %v", failed2, tc.wantFailed)
			}
			if skipped2 != tc.wantSkipped {
				t.Errorf("derived Skipped returned %v but want %v", skipped2, tc.wantSkipped)
			}
		})
	}
}

func TestA_WithContext_cancels_on_cleanup(t *testing.T) {
	var ctx context.Context
	res := goyek.NewRunner(func(a *goyek.A) {
		ctx = a.WithContext(context.Background()).Context()
	})(goyek.Input{})

	if res.Status != goyek.StatusPassed {
		t.Errorf("status was %s but want %s", res.Status, goyek.StatusPassed)
	}

	select {
	case <-ctx.Done():
	default:
		t.Error("context should be canceled after the task finishes")
	}
}

func TestA_WithContext_nil(t *testing.T) {
	out := &strings.Builder{}
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Log("1")
		a.WithContext(nil) //nolint:staticcheck // panic intentionally
		a.Log("2")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: goyek.SyncWriter(out)})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
	assertEqual(t, got.PanicValue, "nil context", "should return proper panic value")
	assertEqual(t, out.String(), "1\n", "should interrupt execution")
}

func TestA_WithContext_concurrent_fail_derived(t *testing.T) {
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	got := goyek.NewRunner(func(a *goyek.A) {
		a2 := a.WithContext(a.Context())
		go func() {
			a2.Fail()
		}()
		for {
			if a.Failed() {
				return
			}
			select {
			case <-timeout.C:
				t.Error("test timeout")
				return
			default:
			}
		}
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
}

func TestA_PostCompletionPanic(t *testing.T) {
	const envKey = "GOYEK_POST_COMPLETION_ENV"
	prevEnv, envWasSet := os.LookupEnv(envKey)
	if err := os.Unsetenv(envKey); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if envWasSet {
			_ = os.Setenv(envKey, prevEnv)
		} else {
			_ = os.Unsetenv(envKey)
		}
	}()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	targetDir := t.TempDir()

	var capturedA, derivedA *goyek.A
	res := goyek.NewRunner(func(a *goyek.A) {
		capturedA = a
		derivedA = a.WithContext(context.Background())
	})(goyek.Input{})

	if res.Status != goyek.StatusPassed {
		t.Fatalf("expected task to pass, got %s", res.Status)
	}

	testCases := []struct {
		name string
		fn   func()
	}{
		{name: "Cleanup", fn: func() { capturedA.Cleanup(func() {}) }},
		{name: "Cleanup derived A", fn: func() { derivedA.Cleanup(func() {}) }},
		{name: "WithContext", fn: func() { capturedA.WithContext(context.Background()) }},
		{name: "Setenv", fn: func() { capturedA.Setenv(envKey, "value") }},
		{name: "TempDir", fn: func() { capturedA.TempDir() }},
		{name: "Chdir", fn: func() { capturedA.Chdir(targetDir) }},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			returned := false
			var panicValue interface{}
			func() {
				defer func() {
					panicValue = recover()
				}()
				tc.fn()
				returned = true
			}()
			if returned {
				t.Fatalf("expected %s after cleanup completion to panic", tc.name)
			}
			method := strings.TrimSuffix(tc.name, " derived A")
			want := method + " called after task cleanup has completed"
			assertEqual(t, panicValue, want, "should panic with the expected value")
		})
	}

	if _, ok := os.LookupEnv(envKey); ok {
		t.Errorf("Setenv changed %s after cleanup completion", envKey)
	}
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if currentDir != originalDir {
		t.Errorf("Chdir changed working directory after cleanup completion: got %q, want %q", currentDir, originalDir)
	}
}

func TestA_Cleanup_concurrent_with_completion(t *testing.T) {
	const attempts = 1000
	const wantPanic = "Cleanup called after task cleanup has completed"
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	for i := 0; i < attempts; i++ {
		start := make(chan struct{})
		cleanupCalled := make(chan struct{}, 1)
		attemptDone := make(chan interface{}, 1)
		runnerDone := make(chan goyek.Result, 1)
		go func() {
			runnerDone <- goyek.NewRunner(func(a *goyek.A) {
				a.Cleanup(func() {
					close(start)
				})
				go func() {
					<-start
					var panicValue interface{}
					func() {
						defer func() {
							panicValue = recover()
						}()
						a.Cleanup(func() {
							cleanupCalled <- struct{}{}
						})
					}()
					attemptDone <- panicValue
				}()
			})(goyek.Input{})
		}()

		var res goyek.Result
		select {
		case res = <-runnerDone:
		case <-timeout.C:
			t.Fatalf("attempt %d: runner did not finish", i)
		}

		if res.Status != goyek.StatusPassed {
			t.Fatalf("attempt %d: expected task to pass, got %s", i, res.Status)
		}
		var panicValue interface{}
		select {
		case panicValue = <-attemptDone:
		case <-timeout.C:
			t.Fatalf("attempt %d: Cleanup call did not finish", i)
		}
		if panicValue != nil {
			assertEqual(t, panicValue, wantPanic, "should reject only after cleanup completion")
			continue
		}
		select {
		case <-cleanupCalled:
		default:
			t.Fatalf("attempt %d: Cleanup returned without running the registered function", i)
		}
	}
}

func TestA_Setenv_concurrent_with_completion(t *testing.T) {
	const (
		attempts  = 1000
		envKey    = "GOYEK_CONCURRENT_COMPLETION_ENV"
		wantPanic = "Setenv called after task cleanup has completed"
	)
	prevEnv, envWasSet := os.LookupEnv(envKey)
	defer func() {
		if envWasSet {
			_ = os.Setenv(envKey, prevEnv)
		} else {
			_ = os.Unsetenv(envKey)
		}
	}()
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	for i := 0; i < attempts; i++ {
		if err := os.Unsetenv(envKey); err != nil {
			t.Fatal(err)
		}
		start := make(chan struct{})
		attemptDone := make(chan interface{}, 1)
		runnerDone := make(chan goyek.Result, 1)
		go func() {
			runnerDone <- goyek.NewRunner(func(a *goyek.A) {
				a.Setenv(envKey, "action")
				a.Cleanup(func() {
					close(start)
				})
				go func() {
					<-start
					var panicValue interface{}
					func() {
						defer func() {
							panicValue = recover()
						}()
						a.Setenv(envKey, "late")
					}()
					attemptDone <- panicValue
				}()
			})(goyek.Input{})
		}()

		var res goyek.Result
		select {
		case res = <-runnerDone:
		case <-timeout.C:
			t.Fatalf("attempt %d: runner did not finish", i)
		}

		if res.Status != goyek.StatusPassed {
			t.Fatalf("attempt %d: expected task to pass, got %s", i, res.Status)
		}
		var panicValue interface{}
		select {
		case panicValue = <-attemptDone:
		case <-timeout.C:
			t.Fatalf("attempt %d: Setenv call did not finish", i)
		}
		if panicValue != nil {
			assertEqual(t, panicValue, wantPanic, "should reject only before changing the environment")
		}
		if value, ok := os.LookupEnv(envKey); ok {
			_ = os.Unsetenv(envKey)
			t.Fatalf("attempt %d: Setenv left %s=%q after cleanup completion", i, envKey, value)
		}
	}
}

type blockingFatalLogger struct {
	entered chan struct{}
	release chan struct{}
}

func (l *blockingFatalLogger) Log(io.Writer, ...interface{}) {}

func (l *blockingFatalLogger) Logf(io.Writer, string, ...interface{}) {}

func (l *blockingFatalLogger) Fatal(io.Writer, ...interface{}) {
	close(l.entered)
	<-l.release
}

func TestA_FailedCallConcurrentWithCompletion(t *testing.T) {
	logger := &blockingFatalLogger{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	defer func() {
		select {
		case <-logger.release:
		default:
			close(logger.release)
		}
	}()
	missingDir := filepath.Join(t.TempDir(), "missing")
	actionReturned := make(chan context.Context, 1)
	runnerDone := make(chan goyek.Result, 1)
	go func() {
		runnerDone <- goyek.NewRunner(func(a *goyek.A) {
			ctx := a.Context()
			go a.Chdir(missingDir)
			<-logger.entered
			actionReturned <- ctx
		})(goyek.Input{Logger: logger})
	}()

	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	var ctx context.Context
	select {
	case ctx = <-actionReturned:
	case <-timeout.C:
		t.Fatal("failed Chdir call did not reach Logger.Fatal")
	}
	select {
	case <-ctx.Done():
	case <-timeout.C:
		t.Fatal("task cleanup did not start")
	}
	observation := time.NewTimer(100 * time.Millisecond)
	defer observation.Stop()
	select {
	case result := <-runnerDone:
		t.Fatalf("runner returned %s before the admitted Chdir call finished", result.Status)
	case <-observation.C:
	}

	close(logger.release)
	select {
	case result := <-runnerDone:
		assertEqual(t, result.Status, goyek.StatusFailed, "should include the admitted Chdir failure")
	case <-timeout.C:
		t.Fatal("runner did not finish after Logger.Fatal was released")
	}
}

func TestA_WithContext_concurrent_fail_original(t *testing.T) {
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	got := goyek.NewRunner(func(a *goyek.A) {
		a2 := a.WithContext(a.Context())
		go func() {
			a.Fail()
		}()
		for {
			if a2.Failed() {
				return
			}
			select {
			case <-timeout.C:
				t.Error("test timeout")
				return
			default:
			}
		}
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
}

func TestA_WithContext_concurrent_cleanup(t *testing.T) {
	out := &strings.Builder{}

	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	var derivedCalled, originalCalled bool

	got := goyek.NewRunner(func(a *goyek.A) {
		a2 := a.WithContext(a.Context())

		ch := make(chan struct{})
		go func() {
			defer func() { close(ch) }()
			a2.Cleanup(func() {
				derivedCalled = true
			})
		}()
		a.Cleanup(func() {
			originalCalled = true
		})
		<-ch
	})(goyek.Input{Output: goyek.SyncWriter(out)})

	assertEqual(t, got.Status, goyek.StatusPassed, "should return proper status")
	assertTrue(t, originalCalled, "original cleanup called")
	assertTrue(t, derivedCalled, "derived cleanup called")
}

func TestA_Cleanup(t *testing.T) {
	out := &strings.Builder{}

	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Cleanup(func() {
				a.Log("5")
				panic("second panic")
			})
			a.Cleanup(func() {
				a.Log("4")
			})
			a.Log("3")
			panic("first panic")
		})
		a.Log("1")
		a.Cleanup(func() {
			a.Log("2")
		})
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: goyek.SyncWriter(out)})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
	assertEqual(t, got.PanicValue, "first panic", "should return proper panic value")
	assertContains(t, out, "1\n2\n3\n4\n5", "should call cleanup funcs in LIFO order")
}

func TestA_Cleanup_when_action_panics(t *testing.T) {
	out := &strings.Builder{}

	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Cleanup(func() {
				a.Log("5")
				panic("second panic")
			})
			a.Cleanup(func() {
				a.Log("4")
			})
			a.Log("3")
			panic("first panic")
		})
		a.Log("1")
		a.Cleanup(func() {
			a.Log("2")
		})
		panic("action panic")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: goyek.SyncWriter(out)})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
	assertEqual(t, got.PanicValue, "action panic", "should return proper panic value")
	assertContains(t, out, "1\n2\n3\n4\n5", "should call cleanup funcs in LIFO order")
}

func TestA_Cleanup_Fail(t *testing.T) {
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Fail()
		})
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
}

func TestA_Cleanup_nil(t *testing.T) {
	out := &strings.Builder{}
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Cleanup(func() {
			a.Log("3")
		})
		a.Log("1")
		a.Cleanup(nil) // nil cleanup func should panic
		a.Log("2")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: goyek.SyncWriter(out)})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
	assertEqual(t, got.PanicValue, "nil cleanup", "should return proper panic value")
	assertEqual(t, out.String(), "1\n3\n", "should interrupt execution but run previously registered cleanups")
}

func TestA_Setenv(t *testing.T) {
	key := "GOYEK_TEST_ENV"
	val := "1"

	res := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv(key, val)

		got := os.Getenv(key)
		assertEqual(t, got, val, "should set the value")
	})(goyek.Input{})

	assertEqual(t, res.Status, goyek.StatusPassed, "should return proper status")
	got := os.Getenv(key)
	assertEqual(t, got, "", "should restore the value after the action")
}

func TestA_Setenv_restore(t *testing.T) {
	key := "GOYEK_TEST_ENV"
	prev := "0"
	val := "1"
	os.Setenv(key, prev)
	defer os.Unsetenv(key)

	res := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv(key, val)

		got := os.Getenv(key)
		assertEqual(t, got, val, "should set the value")
	})(goyek.Input{})

	assertEqual(t, res.Status, goyek.StatusPassed, "should return proper status")
	got := os.Getenv(key)
	assertEqual(t, got, prev, "should restore the value after the action")
}

func TestA_TempDir(t *testing.T) {
	var dir string
	res := goyek.NewRunner(func(a *goyek.A) {
		dir = a.TempDir()

		_, err := os.Lstat(dir)
		assertEqual(t, err, nil, "the dir should exixt")
	})(goyek.Input{TaskName: "0!ą😊"})

	assertEqual(t, res.Status, goyek.StatusPassed, "should return proper status")
	_, err := os.Lstat(dir)
	assertTrue(t, os.IsNotExist(err), "should remove the dir after the action")
}

func TestA_TempDir_error(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("TMPDIR is not used on this platform")
	}

	missingTempDir := filepath.Join(t.TempDir(), "missing")
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv("TMPDIR", missingTempDir)
		a.TempDir()
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should fail when the temporary directory cannot be created")
}

func requireModePermissionErrors(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("directory permissions differ on this platform")
	}

	probeDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(probeDir, "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	originalDir, err := os.Open(".")
	if err != nil {
		t.Fatal(err)
	}
	probeFD, err := os.Open(probeDir) //nolint:gosec // probeDir is created by t.TempDir
	if err != nil {
		_ = originalDir.Close()
		t.Fatal(err)
	}
	defer func() {
		_ = originalDir.Chdir()
		_ = originalDir.Close()
		_ = probeFD.Close()
		_ = os.Chmod(probeDir, 0o700) //nolint:gosec // restore directory access for cleanup
		_ = os.RemoveAll(probeDir)
	}()

	if err := os.Chmod(probeDir, 0); err != nil {
		t.Fatal(err)
	}
	openedDir, openErr := os.Open(probeDir) //nolint:gosec // probeDir is created by t.TempDir
	if openedDir != nil {
		_ = openedDir.Close()
	}
	chdirErr := probeFD.Chdir()
	if chdirErr == nil {
		if err := originalDir.Chdir(); err != nil {
			t.Fatalf("restore working directory after permission probe: %v", err)
		}
	}
	removeErr := os.RemoveAll(probeDir)
	if openErr == nil || chdirErr == nil || removeErr == nil {
		t.Skip("process can bypass directory mode permissions")
	}
}

func TestA_TempDir_cleanup_error(t *testing.T) {
	requireModePermissionErrors(t)

	var (
		dir      string
		setupErr error
	)
	got := goyek.NewRunner(func(a *goyek.A) {
		dir = a.TempDir()
		if setupErr = os.Mkdir(filepath.Join(dir, "child"), 0o700); setupErr != nil {
			a.Fatal(setupErr)
		}
		if setupErr = os.Chmod(dir, 0); setupErr != nil {
			a.Fatal(setupErr)
		}
	})(goyek.Input{})
	defer func() {
		_ = os.Chmod(dir, 0o700) //nolint:gosec // restore directory access for cleanup
		_ = os.RemoveAll(dir)
	}()
	if setupErr != nil {
		t.Fatal(setupErr)
	}

	assertEqual(t, got.Status, goyek.StatusFailed, "should report a temporary directory cleanup failure")
}

func TestA_TempDir_UTF8SafeTruncation(t *testing.T) {
	testCases := []struct {
		name     string
		taskName string
	}{
		{
			name:     "2-byte truncation",
			taskName: strings.Repeat("a", 63) + "ą",
		},
		{
			name:     "4-byte truncation",
			taskName: strings.Repeat("a", 63) + "💩",
		},
		{
			name:     "4-byte truncation middle",
			taskName: strings.Repeat("a", 61) + "💩",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var dir string
			res := goyek.NewRunner(func(a *goyek.A) {
				dir = a.TempDir()
			})(goyek.Input{TaskName: tc.taskName})

			assertEqual(t, res.Status, goyek.StatusPassed, "should return proper status")
			base := filepath.Base(dir)
			if !utf8.ValidString(base) {
				t.Errorf("TempDir name is not valid UTF-8: %q", base)
			}
		})
	}
}

func TestA_TempDir_long_name(t *testing.T) {
	var dir string
	res := goyek.NewRunner(func(a *goyek.A) {
		dir = a.TempDir()

		_, err := os.Lstat(dir)
		assertEqual(t, err, nil, "the dir should exist")
	})(goyek.Input{TaskName: strings.Repeat("a", 300)})

	assertEqual(t, res.Status, goyek.StatusPassed, "should return proper status")
	_, err := os.Lstat(dir)
	assertTrue(t, os.IsNotExist(err), "should remove the dir after the action")
}

func TestA_Chdir(t *testing.T) {
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir) //nolint:errcheck // not checking errors for cleanup

	// The "relative" test case relies on tmp not being a symlink.
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(oldDir, tmp)
	if err != nil {
		// If GOROOT is on C: volume and tmp is on the D: volume, there
		// is no relative path between them, so skip that test case.
		rel = "skip"
	}

	for _, tc := range []struct {
		name, dir, pwd string
		extraChdir     bool
	}{
		{
			name: "absolute",
			dir:  tmp,
			pwd:  tmp,
		},
		{
			name: "relative",
			dir:  rel,
			pwd:  tmp,
		},
		{
			name: "current (absolute)",
			dir:  oldDir,
			pwd:  oldDir,
		},
		{
			name: "current (relative) with extra os.Chdir",
			dir:  ".",
			pwd:  oldDir,

			extraChdir: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.dir == "skip" {
				t.Skipf("skipping test because there is no relative path between %s and %s", oldDir, tmp)
			}
			if !filepath.IsAbs(tc.pwd) {
				t.Fatalf("Bad tc.pwd: %q (must be absolute)", tc.pwd)
			}

			res := goyek.NewRunner(func(a *goyek.A) {
				a.Chdir(tc.dir)

				newDir, err := os.Getwd()
				if err != nil {
					t.Error(err)
					return
				}
				if newDir != tc.pwd {
					t.Errorf("failed to chdir to %q: getwd: got %q, want %q", tc.dir, newDir, tc.pwd)
					return
				}

				switch runtime.GOOS {
				case "windows", "plan9":
					// Windows and Plan 9 do not use the PWD variable.
				default:
					if pwd := os.Getenv("PWD"); pwd != tc.pwd {
						t.Errorf("PWD: got %q, want %q", pwd, tc.pwd)
						return
					}
				}

				if tc.extraChdir {
					_ = os.Chdir("..")
				}
			})(goyek.Input{})
			assertEqual(t, res.Status, goyek.StatusPassed, "should return proper status")

			newDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if newDir != oldDir {
				t.Fatalf("failed to restore wd to %s: getwd: %s", oldDir, newDir)
			}
		})
	}
}

func TestA_Chdir_open_error(t *testing.T) {
	requireModePermissionErrors(t)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	lockedDir := t.TempDir()
	targetDir := t.TempDir()
	if err := os.Chdir(lockedDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(lockedDir, 0o700) //nolint:gosec // restore directory access for cleanup
		_ = os.Chdir(originalDir)
	}()
	if err := os.Chmod(lockedDir, 0); err != nil {
		t.Fatal(err)
	}

	got := goyek.NewRunner(func(a *goyek.A) {
		a.Chdir(targetDir)
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should fail when the current directory cannot be opened")
}

func TestA_Chdir_restore_error(t *testing.T) {
	requireModePermissionErrors(t)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	sourceDir := t.TempDir()
	targetDir := t.TempDir()
	if err := os.Chdir(sourceDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(sourceDir, 0o700) //nolint:gosec // restore directory access for cleanup
		_ = os.Chdir(originalDir)
	}()

	got := goyek.NewRunner(func(a *goyek.A) {
		a.Chdir(targetDir)
		if err := os.Chmod(sourceDir, 0); err != nil {
			a.Fatal(err)
		}
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should fail when the working directory cannot be restored")
	panicValue, ok := got.PanicValue.(string)
	if !ok || !strings.Contains(panicValue, "goyek.Chdir:") {
		t.Errorf("expected a Chdir restoration panic, got %#v", got.PanicValue)
	}
}

func TestA_uses_Logger_dynamic_interface(t *testing.T) {
	testCases := []struct {
		desc   string
		action func(a *goyek.A)
	}{
		{
			desc:   "Helper",
			action: func(a *goyek.A) { a.Helper() },
		},
		{
			desc:   "Log",
			action: func(a *goyek.A) { a.Log() },
		},
		{
			desc:   "Logf",
			action: func(a *goyek.A) { a.Logf("") },
		},
		{
			desc:   "Error",
			action: func(a *goyek.A) { a.Error() },
		},
		{
			desc:   "Errorf",
			action: func(a *goyek.A) { a.Errorf("") },
		},
		{
			desc:   "Fatal",
			action: func(a *goyek.A) { a.Fatal() },
		},
		{
			desc:   "Fatalf",
			action: func(a *goyek.A) { a.Fatalf("") },
		},
		{
			desc:   "Helper",
			action: func(a *goyek.A) { a.Skip() },
		},
		{
			desc:   "Skipf",
			action: func(a *goyek.A) { a.Skipf("") },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			flow := &goyek.Flow{}

			flow.SetOutput(io.Discard)
			loggerSpy := &helperLoggerSpy{}
			flow.SetLogger(loggerSpy)
			flow.Define(goyek.Task{
				Name:   "task",
				Action: tc.action,
			})

			_ = flow.Execute(context.Background(), []string{"task"})

			assertTrue(t, loggerSpy.called, "called logger")
		})
	}
}

type helperLoggerSpy struct {
	called bool
}

func (l *helperLoggerSpy) Log(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Logf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Error(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Errorf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Fatal(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Fatalf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Skip(_ io.Writer, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Skipf(_ io.Writer, _ string, _ ...interface{}) {
	l.called = true
}

func (l *helperLoggerSpy) Helper() {
	l.called = true
}

func TestA_Setenv_parallel_panic(t *testing.T) {
	out := &strings.Builder{}
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv("KEY", "VALUE")
	})(goyek.Input{Parallel: true, Output: goyek.SyncWriter(out), Logger: goyek.FmtLogger{}})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
	assertContains(t, out, "Setenv called in a parallel task", "should log error message")
}

func TestA_Chdir_parallel_panic(t *testing.T) {
	out := &strings.Builder{}
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Chdir(".")
	})(goyek.Input{Parallel: true, Output: goyek.SyncWriter(out), Logger: goyek.FmtLogger{}})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
	assertContains(t, out, "Chdir called in a parallel task", "should log error message")
}

func TestA_Setenv_error(t *testing.T) {
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Setenv("", "")
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
}

func TestA_Chdir_error(t *testing.T) {
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Chdir("non-existent-directory-@!#$")
	})(goyek.Input{})

	assertEqual(t, got.Status, goyek.StatusFailed, "should return proper status")
}
