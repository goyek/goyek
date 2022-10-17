package main

import "github.com/goyek/goyek/v2"

var golint = flow.Define(goyek.Task{
	Name:  "golint",
	Usage: "golangci-lint run --fix",
	Action: func(tf *goyek.TF) {
		Exec(tf, dirRoot, "golangci-lint run --fix")
		Exec(tf, dirBuild, "golangci-lint run --fix")
	},
})
