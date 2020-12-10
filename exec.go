package taskflow

import (
	"os"
	"os/exec"
	"strings"
)

// Exec runs the specified command and waits for it to complete.
// The process stderr and stdout are piped to the Output.
// Use workdir argument to change the process working directory.
// Use env argument to provide additional environment variables in the form "key=value".
func (tf *TF) Exec(workdir string, env []string, name string, args ...string) error {
	cmdStr := strings.Join(append([]string{name}, args...), " ")
	tf.Logf("Exec: %s", cmdStr)

	cmd := exec.CommandContext(tf.Context(), name, args...) //nolint:gosec // yes, this runs a subprocess
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stderr = tf.Output()
	cmd.Stdout = tf.Output()
	return cmd.Run()
}
