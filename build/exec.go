package main

import (
	"strings"

	"github.com/goyek/goyek"
)

// Exec runs the command in given directory.
func Exec(tf *goyek.TF, workDir, cmdLine string) error {
	tf.Logf("Run '%s' in %s", cmdLine, workDir)
	args := strings.Split(cmdLine, " ")
	cmd := tf.Cmd(args[0], args[1:]...)
	cmd.Dir = workDir
	return cmd.Run()
}
