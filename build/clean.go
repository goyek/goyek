package main

import "github.com/goyek/goyek/v2"

var clean = flow.Define(goyek.Task{
	Name:  "clean",
	Usage: "remove git ignored files",
	Action: func(tf *goyek.TF) {
		Exec(tf, rootDir, "git clean -fXd")
	},
})
