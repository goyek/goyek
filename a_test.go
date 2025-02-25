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

	"github.com/goyek/goyek/v2"
)

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
			})(goyek.Input{Context: ctx, Output: sb, Logger: goyek.FmtLogger{}})

			if res.Status != tc.wantStatus {
				t.Errorf("status was %s but want %s", res.Status, tc.wantStatus)
			}
			if got != ctx {
				t.Errorf("original Context returned %v but want %v", got, ctx)
			}
			if got2 != newCtx {
				t.Errorf("derived Context returned %v but want %v", got2, newCtx)
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

func TestA_WithContext_nil(t *testing.T) {
	out := &strings.Builder{}
	got := goyek.NewRunner(func(a *goyek.A) {
		a.Log("1")
		a.WithContext(nil) //nolint:staticcheck // panic intentionally
		a.Log("2")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

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
	})(goyek.Input{Output: out})

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
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

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
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

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
		a.Cleanup(nil) // nil cleanup func is gracefully ignored
		a.Log("2")
	})(goyek.Input{Logger: &goyek.FmtLogger{}, Output: out})

	assertEqual(t, got.Status, goyek.StatusPassed, "should return proper status")
	assertEqual(t, out.String(), "1\n2\n3\n", "should continue execution")
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
