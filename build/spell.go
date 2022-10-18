package main

import (
	"strings"

	"github.com/goyek/goyek/v2"
)

var spell = flow.Define(goyek.Task{
	Name:  "spell",
	Usage: "misspell",
	Action: func(tf *goyek.TF) {
		if ok := Exec(tf, dirBuild, "go install github.com/client9/misspell/cmd/misspell"); !ok {
			return
		}
		mdFiles := find(tf, ".md")
		if len(mdFiles) == 0 {
			tf.Skip("no .md files")
		}
		Exec(tf, dirRoot, "misspell -error -locale=US -w "+strings.Join(mdFiles, " "))
	},
})
