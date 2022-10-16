package main

import "github.com/goyek/goyek/v2"

var spell = flow.Define(goyek.Task{
	Name:  "spell",
	Usage: "misspell",
	Action: func(tf *goyek.TF) {
		Exec(tf, dirRoot, "misspell -error -locale=US -i=importas -w .")
	},
})
