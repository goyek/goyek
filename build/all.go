package main

import "github.com/goyek/goyek/v2"

var all = goyek.Define(goyek.Task{
	Name:  "all",
	Usage: "build pipeline",
	Deps: goyek.Deps{
		mod,
		lint,
		test,
	},
})
