package middleware

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestBufferedMiddleware_removesSpillFile(t *testing.T) {
	tempDir := t.TempDir()
	useSpoolTempDir(t, tempDir)
	message := strings.Repeat("x", maxBufferedOutputBytes+1)

	tests := []struct {
		name      string
		wrap      func(goyek.Runner) goyek.Runner
		status    goyek.Status
		panic     bool
		wantBytes int64
	}{
		{
			name:      "buffer parallel passed",
			wrap:      BufferParallel,
			status:    goyek.StatusPassed,
			wantBytes: int64(len(message)),
		},
		{
			name:      "buffer parallel failed",
			wrap:      BufferParallel,
			status:    goyek.StatusFailed,
			wantBytes: int64(len(message)),
		},
		{
			name:  "buffer parallel panic",
			wrap:  BufferParallel,
			panic: true,
		},
		{
			name:   "silent passed",
			wrap:   SilentNonFailed,
			status: goyek.StatusPassed,
		},
		{
			name:      "silent failed",
			wrap:      SilentNonFailed,
			status:    goyek.StatusFailed,
			wantBytes: int64(len(message)),
		},
		{
			name:  "silent panic",
			wrap:  SilentNonFailed,
			panic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &middlewareCountWriter{}
			runner := tt.wrap(func(in goyek.Input) goyek.Result {
				if n, err := io.WriteString(in.Output, message); err != nil || n != len(message) {
					t.Fatalf("WriteString returned %d, %v; want %d, nil", n, err, len(message))
				}
				if tt.panic {
					panic("boom")
				}
				return goyek.Result{Status: tt.status}
			})

			var recovered interface{}
			func() {
				defer func() {
					recovered = recover()
				}()
				runner(goyek.Input{
					Parallel: true,
					Output:   goyek.SyncWriter(output),
				})
			}()

			if tt.panic && recovered == nil {
				t.Fatal("runner did not propagate panic")
			}
			if !tt.panic && recovered != nil {
				t.Fatalf("runner unexpectedly panicked: %v", recovered)
			}
			if output.n != tt.wantBytes {
				t.Fatalf("emitted %d bytes, want %d", output.n, tt.wantBytes)
			}
			matches, err := filepath.Glob(filepath.Join(tempDir, "goyek-output-*"))
			if err != nil {
				t.Fatalf("glob spill files: %v", err)
			}
			if len(matches) != 0 {
				t.Fatalf("spill files remain after runner exit: %v", matches)
			}
		})
	}
}

type middlewareCountWriter struct {
	n int64
}

func (w *middlewareCountWriter) Write(p []byte) (int, error) {
	w.n += int64(len(p))
	return len(p), nil
}

func useSpoolTempDir(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "plan9" {
		t.Skip("Plan 9 does not select the temporary directory from the environment")
	}

	names := []string{"TMPDIR"}
	if runtime.GOOS == "windows" {
		names = []string{"TMP", "TEMP"}
	}
	for _, name := range names {
		name := name
		old, existed := os.LookupEnv(name)
		if err := os.Setenv(name, dir); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(name, old)
			} else {
				_ = os.Unsetenv(name)
			}
		})
	}
}
