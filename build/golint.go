package main

import "github.com/goyek/goyek/v3"

var golint = goyek.Define(goyek.Task{
	Name:  "golint",
	Usage: "golangci-lint run --fix",
	Action: func(a *goyek.A) {
		if !ExecArgs(a, dirBuild, "go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint") {
			return
		}
		ExecArgs(a, dirRoot, "golangci-lint", "run", "--fix")
		ExecArgs(a, dirBuild, "golangci-lint", "run", "--fix")
	},
})
