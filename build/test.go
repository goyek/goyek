package main

import "github.com/goyek/goyek/v2"

var test = goyek.Define(goyek.Task{
	Name:  "test",
	Usage: "go test",
	Action: func(tf *goyek.TF) {
		verbose := ""
		if *v {
			verbose = "-v"
		}
		if !Exec(tf, dirRoot, "go test "+verbose+" -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...") {
			return
		}
		Exec(tf, dirRoot, "go tool cover -html=coverage.out -o coverage.html")
	},
})
