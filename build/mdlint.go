package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/goyek/goyek/v2"
)

var mdlint = goyek.Define(goyek.Task{
	Name:  "mdlint",
	Usage: "markdownlint-cli (uses docker)",
	Action: func(tf *goyek.TF) {
		if _, err := exec.LookPath("docker"); err != nil {
			tf.Skip(err)
		}
		curDir, err := os.Getwd()
		if err != nil {
			tf.Fatal(err)
		}
		mdFiles := find(tf, ".md")
		if len(mdFiles) == 0 {
			tf.Skip("no .md files")
		}
		dockerImage := "ghcr.io/igorshubovych/markdownlint-cli:v0.32.2"
		Exec(tf, dirRoot, "docker run --rm -v '"+curDir+":/workdir' "+dockerImage+" "+strings.Join(mdFiles, " "))
	},
})
