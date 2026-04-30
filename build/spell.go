package main

import (
	"github.com/goyek/goyek/v3"
)

var spell = goyek.Define(goyek.Task{
	Name:  "spell",
	Usage: "misspell",
	Action: func(a *goyek.A) {
		if !Exec(a, dirBuild, "go", "install", "github.com/client9/misspell/cmd/misspell") {
			return
		}
		mdFiles := find(a, ".md")
		if len(mdFiles) == 0 {
			a.Skip("no .md files")
		}
		args := []string{"-error", "-locale=US", "-w"}
		args = append(args, mdFiles...)
		Exec(a, dirRoot, "misspell", args...)
	},
})
