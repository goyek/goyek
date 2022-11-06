package main

import "github.com/goyek/goyek/v2"

var mod = goyek.Define(goyek.Task{
	Name:  "mod",
	Usage: "go mod tidy",
	Action: func(a *goyek.A) {
		Exec(a, dirRoot, "go mod tidy")
		Exec(a, dirBuild, "go mod tidy")
	},
})
