package main

import (
	"io"
	"os/exec"
	"strings"

	"github.com/goyek/goyek/v2"
)

var diff = goyek.Define(goyek.Task{
	Name:  "diff",
	Usage: "git diff",
	Action: func(a *goyek.A) {
		Exec(a, dirRoot, "git diff --exit-code")

		a.Log("Cmd: git status --porcelain")

		sb := &strings.Builder{}
		cmd := exec.CommandContext(a.Context(), "git", "status", "--porcelain")
		cmd.Stdout = io.MultiWriter(a.Output(), sb)
		cmd.Stderr = io.MultiWriter(a.Output(), sb)
		if err := cmd.Run(); err != nil {
			a.Error(err)
		}
		if sb.Len() > 0 {
			a.Error("git status --porcelain returned output")
		}
	},
})
