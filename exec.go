package taskflow

import (
	"os/exec"
	"strings"
)

func Exec(tf *TF, workdir string, env []string, name string, args ...string) error {
	cmd := exec.CommandContext(tf.Context(), name, args...) //nolint:gosec // yes, this runs a subprocess
	cmd.Dir = workdir
	cmd.Env = env
	cmd.Stderr = tf.Writer()
	cmd.Stdout = tf.Writer()
	cmdStr := strings.Join(append([]string{name}, args...), " ")
	tf.Logf("Exec: %s", cmdStr)
	return cmd.Run()
}
