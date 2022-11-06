package main

import (
	"github.com/mattn/go-shellwords"

	"github.com/goyek/goyek/v2"
)

// Exec runs the command in given directory.
// It calls a.Error[f] and returns false in case of any problems.
func Exec(a *goyek.A, workDir, cmdLine string) bool {
	a.Helper()
	a.Logf("Run %q in %s", cmdLine, workDir)
	args, err := shellwords.Parse(cmdLine)
	if err != nil {
		a.Errorf("parse command line: %v", err)
		return false
	}
	cmd := a.Cmd(args[0], args[1:]...)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		a.Error(err)
		return false
	}
	return true
}
