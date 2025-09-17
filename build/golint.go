package main

import "github.com/goyek/goyek/v2"

var golint = goyek.Define(goyek.Task{
	Name:  "golint",
	Usage: "golangci-lint run --fix",
	Action: func(a *goyek.A) {
		if !Exec(a, dirBuild, "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint") {
			return
		}
		Exec(a, dirRoot, "golangci-lint run --fix")
		Exec(a, dirBuild, "golangci-lint run --fix")
	},
})
