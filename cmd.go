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
