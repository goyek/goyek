package goyek

import (
	"os/exec"
	"strings"
)

// Cmd is like exec.Command, but it assigns a's context
// and assigns Stdout and Stderr to a's output.
func (a *A) Cmd(name string, args ...string) *exec.Cmd {
	cmdStr := strings.Join(append([]string{name}, args...), " ")
	a.Logf("Cmd: %s", cmdStr)

	cmd := exec.CommandContext(a.Context(), name, args...) //nolint:gosec // yes, this runs a subprocess
	cmd.Stderr = a.Output()
	cmd.Stdout = a.Output()
	return cmd
}

// Exec returns a action that will run the named program with the given arguments.
// The action will pass only if the program if the program runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit status.
func Exec(name string, args ...string) func(*A) {
	return func(a *A) {
		if err := a.Cmd(name, args...).Run(); err != nil {
			a.Fatalf("Cmd %s failed: %v", name, err)
		}
	}
}
