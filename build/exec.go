package main

import (
	"strings"

	"github.com/goyek/goyek/v2"
)

// Exec runs the command in given directory.
// Returns true if it finished with 0 exit code.
// Otherwise, reports error and returns false .
func Exec(tf *goyek.TF, workDir, cmdLine string) {
	tf.Logf("Run %q in %s", cmdLine, workDir)
	args := strings.Split(cmdLine, " ")
	cmd := tf.Cmd(args[0], args[1:]...)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		tf.Error(err)
	}
}
