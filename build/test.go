package main

import "github.com/goyek/goyek/v2"

var test = flow.Define(goyek.Task{
	Name:  "test",
	Usage: "go test",
	Action: func(tf *goyek.TF) {
		Exec(tf, rootDir, "go test -race -covermode=atomic -coverprofile=coverage.out ./...")
	},
})
