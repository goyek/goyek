package main

import "github.com/goyek/goyek/v2"

var golint = flow.Define(goyek.Task{
	Name:  "golint",
	Usage: "golangci-lint run --fix",
	Action: func(tf *goyek.TF) {
		if ok := Exec(tf, dirBuild, "go install github.com/golangci/golangci-lint/cmd/golangci-lint"); !ok {
			return
		}
		Exec(tf, dirRoot, "golangci-lint run --fix")
		Exec(tf, dirBuild, "golangci-lint run --fix")
	},
})
