package main

import "github.com/goyek/goyek/v2"

var test = goyek.Define(goyek.Task{
	Name:  "test",
	Usage: "go test",
	Action: func(a *goyek.A) {
		verbose := ""
		if *v {
			verbose = "-v"
		}
		if !Exec(a, dirRoot, "go test "+verbose+" -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...") {
			return
		}
		Exec(a, dirRoot, "go tool cover -html=coverage.out -o coverage.html")
	},
})
