package main

import "github.com/goyek/goyek/v3"

var lint = goyek.Define(goyek.Task{
	Name:  "lint",
	Usage: "all linters",
	Deps: goyek.Deps{
		golint,
		spell,
		mdlint,
	},
})
