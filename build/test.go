package main

import "github.com/goyek/goyek/v2"

var test = flow.Define(goyek.Task{
	Name:  "test",
	Usage: "go test",
	Action: func(tf *goyek.TF) {
		verbose := ""
		if *v {
			verbose = "-v"
		}
		Exec(tf, dirRoot, "go test "+verbose+" -race -covermode=atomic -coverprofile=coverage.out ./...")
	},
})
