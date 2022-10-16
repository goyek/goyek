package main

import (
	"strings"

	"github.com/goyek/goyek/v2"
)

var install = flow.Define(goyek.Task{
	Name:  "install",
	Usage: "go install tools",
	Action: func(tf *goyek.TF) {
		tools := &strings.Builder{}
		toolsCmd := tf.Cmd("go", "list", `-f={{ join .Imports " " }}`, "-tags=tools")
		toolsCmd.Dir = dirTools
		toolsCmd.Stdout = tools
		if err := toolsCmd.Run(); err != nil {
			tf.Fatal(err)
		}

		Exec(tf, dirTools, "go install "+strings.TrimSpace(tools.String()))
	},
})
