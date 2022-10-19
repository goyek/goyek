package main

import "github.com/goyek/goyek/v2"

var _ = goyek.Define(goyek.Task{
	Name:  "clean",
	Usage: "remove git ignored files",
	Action: func(tf *goyek.TF) {
		Exec(tf, dirRoot, "git clean -fXd")
	},
})
