package main

import "github.com/goyek/goyek/v2"

var mod = flow.Define(goyek.Task{
	Name:  "mod",
	Usage: "go mod tidy",
	Action: func(tf *goyek.TF) {
		Exec(tf, rootDir, "go mod tidy")
		Exec(tf, buildDir, "go mod tidy")
		Exec(tf, toolsDir, "go mod tidy")
	},
})
