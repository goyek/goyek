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

func TestSilentNonFailed_removesSpillFile(t *testing.T) {
	tempDir := t.TempDir()
	setProcessTempDir(t, tempDir)
	message := strings.Repeat("x", maxBufferedOutputBytes+1)

	tests := []struct {
		name      string
		status    goyek.Status
		panic     bool
		wantBytes int64
	}{
		{name: "passed", status: goyek.StatusPassed},
		{name: "failed", status: goyek.StatusFailed, wantBytes: int64(len(message))},
		{name: "panic", panic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &countWriter{}
			runner := SilentNonFailed(func(in goyek.Input) goyek.Result {
				if _, err := io.WriteString(in.Output, message); err != nil {
					t.Fatalf("writing output: %v", err)
				}
				if tt.panic {
					panic("boom")
				}
				return goyek.Result{Status: tt.status}
			})

			var recovered interface{}
			func() {
				defer func() { recovered = recover() }()
				runner(goyek.Input{Output: goyek.SyncWriter(output)})
			}()

			if tt.panic && recovered == nil {
				t.Fatal("runner did not propagate panic")
			}
			if !tt.panic && recovered != nil {
				t.Fatalf("runner unexpectedly panicked: %v", recovered)
			}
			if output.n != tt.wantBytes {
				t.Fatalf("replayed %d bytes, want %d", output.n, tt.wantBytes)
			}
			matches, err := filepath.Glob(filepath.Join(tempDir, outputSpillPattern))
			if err != nil {
				t.Fatalf("glob spill files: %v", err)
			}
			if len(matches) != 0 {
				t.Fatalf("spill files remain after runner exit: %v", matches)
			}
		})
	}
}

type countWriter struct {
	n int64
}

func (w *countWriter) Write(p []byte) (int, error) {
	w.n += int64(len(p))
	return len(p), nil
}

func setProcessTempDir(t *testing.T, dir string) {
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
