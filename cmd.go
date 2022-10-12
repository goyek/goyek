package goyek

import (
	"os"
	"os/exec"
)

// Cmd is like exec.Command, but it assigns tf's context
// and assigns Stdout and Stderr to tf's output,
// and Stdin to os.Stdin.
func (tf *TF) Cmd(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(tf.Context(), name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = tf.Output()
	cmd.Stdout = tf.Output()
	return cmd
}
