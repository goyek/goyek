package goyek

import (
	"os/exec"
	"strings"
)

// Cmd is like exec.Command, but it assigns a's context
// and assigns Stdout and Stderr to a's output.
func (p *Progress) Cmd(name string, args ...string) *exec.Cmd {
	cmdStr := strings.Join(append([]string{name}, args...), " ")
	p.Logf("Cmd: %s", cmdStr)

	cmd := exec.CommandContext(p.Context(), name, args...) //nolint:gosec // yes, this runs a subprocess
	cmd.Stderr = p.Output()
	cmd.Stdout = p.Output()
	return cmd
}
