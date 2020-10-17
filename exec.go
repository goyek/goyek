package taskflow

import "os/exec"

func Exec(tf *TF, name string, args ...string) error {
	cmd := exec.CommandContext(tf.Context(), name, args...) //nolint:gosec // yes, this runs a subprocess
	cmd.Stderr = tf.Writer()
	cmd.Stdout = tf.Writer()
	return cmd.Run()
}
