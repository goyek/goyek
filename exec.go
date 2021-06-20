package goyek

import (
	"os/exec"
	"strings"
)

// Cmd is like exec.Command, but it assigns tf's context
// and assigns Stdout and Stderr to tf's output.
func (tf *TF) Cmd(name string, args ...string) *exec.Cmd {
	cmdStr := strings.Join(append([]string{name}, args...), " ")
	tf.Logf("Cmd: %s", cmdStr)

	cmd := exec.CommandContext(tf.Context(), name, args...) //nolint:gosec // yes, this runs a subprocess
	cmd.Stderr = tf.Output()
	cmd.Stdout = tf.Output()
	return cmd
}

// Exec returns a action that will run the named program with the given arguments.
// The action will pass only if the program if the program runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit status.
func Exec(name string, args ...string) func(*TF) {
	return func(tf *TF) {
		if err := tf.Cmd(name, args...).Run(); err != nil {
			tf.Fatalf("Cmd %s failed: %v", name, err)
		}
	}
}
