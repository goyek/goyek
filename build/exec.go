package main

import (
	"os"
	"os/exec"

	"github.com/goyek/goyek/v3"
)

// Exec runs the command in given directory.
// It calls a.Error[f] and returns false in case of any problems.
func Exec(a *goyek.A, workDir, name string, args ...string) bool {
	a.Helper()
	a.Logf("Run %v in %s", append([]string{name}, args...), workDir)
	cmd := exec.CommandContext(a.Context(), name, args...) //nolint:gosec // it is a convenient function to run programs
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = a.Output()
	cmd.Stderr = a.Output()
	if err := cmd.Run(); err != nil {
		a.Error(err)
		return false
	}
	return true
}
