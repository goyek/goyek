package taskflow

import (
	"os"
	"os/exec"
	"strings"
)

func (tf *TF) Exec(workdir string, env []string, name string, args ...string) error {
	cmd := exec.CommandContext(tf.Context(), name, args...) //nolint:gosec // yes, this runs a subprocess
	cmd.Dir = workdir

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)

	cmd.Stderr = tf.Writer()
	cmd.Stdout = tf.Writer()

	cmdStr := strings.Join(append([]string{name}, args...), " ")
	tf.Logf("Exec: %s", cmdStr)
	return cmd.Run()
}
