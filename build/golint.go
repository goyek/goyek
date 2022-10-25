package main

import "github.com/goyek/goyek/v2"

var golint = goyek.Define(goyek.Task{
	Name:  "golint",
	Usage: "golangci-lint run --fix",
	Action: func(tf *goyek.TF) {
		if !Exec(tf, dirBuild, "go install github.com/golangci/golangci-lint/cmd/golangci-lint") {
			return
		}
		Exec(tf, dirRoot, "golangci-lint run --fix")
		Exec(tf, dirBuild, "golangci-lint run --fix")
	},
})
