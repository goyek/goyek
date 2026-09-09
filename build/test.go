package main

import "github.com/goyek/goyek/v3"

const testCommand = "test"

var test = goyek.Define(goyek.Task{
	Name:  testCommand,
	Usage: "go test",
	Action: func(a *goyek.A) {
		args := []string{testCommand}
		if *v {
			args = append(args, "-v")
		}
		args = append(args, "-race", "-covermode=atomic", "-coverprofile=coverage.out", "-coverpkg=./...", "./...")
		if !Exec(a, dirRoot, "go", args...) {
			return
		}
		buildArgs := []string{testCommand}
		if *v {
			buildArgs = append(buildArgs, "-v")
		}
		buildArgs = append(buildArgs, "-race", "./...")
		if !Exec(a, dirBuild, "go", buildArgs...) {
			return
		}
		Exec(a, dirRoot, "go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	},
})
