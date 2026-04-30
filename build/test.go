package main

import "github.com/goyek/goyek/v3"

var test = goyek.Define(goyek.Task{
	Name:  "test",
	Usage: "go test",
	Action: func(a *goyek.A) {
		args := []string{"test"}
		if *v {
			args = append(args, "-v")
		}
		args = append(args, "-race", "-covermode=atomic", "-coverprofile=coverage.out", "-coverpkg=./...", "./...")
		if !Exec(a, dirRoot, "go", args...) {
			return
		}
		Exec(a, dirRoot, "go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	},
})
