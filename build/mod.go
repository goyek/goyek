package main

import "github.com/goyek/goyek/v2"

var mod = goyek.Define(goyek.Task{
	Name:  "mod",
	Usage: "go mod tidy",
	Action: func(tf *goyek.TF) {
		Exec(tf, dirRoot, "go mod tidy")
		Exec(tf, dirBuild, "go mod tidy")
	},
})
