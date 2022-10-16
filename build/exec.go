package main

import (
	"github.com/mattn/go-shellwords"

	"github.com/goyek/goyek/v2"
)

// Exec runs the command in given directory.
// It calls tf.Error[f] in case of any problems.
func Exec(tf *goyek.TF, workDir, cmdLine string) {
	tf.Logf("Run %q in %s", cmdLine, workDir)
	args, err := shellwords.Parse(cmdLine)
	if err != nil {
		tf.Errorf("parse command line: %v", err)
		return
	}
	cmd := tf.Cmd(args[0], args[1:]...)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		tf.Error(err)
	}
}
