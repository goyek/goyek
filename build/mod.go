package main

import "github.com/goyek/goyek/v3"

var mod = goyek.Define(goyek.Task{
	Name:  "mod",
	Usage: "go mod tidy",
	Action: func(a *goyek.A) {
		ExecArgs(a, dirRoot, "go", "mod", "tidy")
		ExecArgs(a, dirBuild, "go", "mod", "tidy")
	},
})
