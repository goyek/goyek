package main

import (
	"io"
	"strings"

	"github.com/goyek/goyek/v2"
)

var diff = goyek.Define(goyek.Task{
	Name:  "diff",
	Usage: "git diff",
	Action: func(tf *goyek.TF) {
		Exec(tf, dirRoot, "git diff --exit-code")

		tf.Log("Cmd: git status --porcelain")
		cmd := tf.Cmd("git", "status", "--porcelain")
		sb := &strings.Builder{}
		cmd.Stdout = io.MultiWriter(tf.Output(), sb)
		if err := cmd.Run(); err != nil {
			tf.Error(err)
		}
		if sb.Len() > 0 {
			tf.Error("git status --porcelain returned output")
		}
	},
})
