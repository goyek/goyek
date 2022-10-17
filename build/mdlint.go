package main

import (
	"os"
	"os/exec"

	"github.com/goyek/goyek/v2"
)

var mdlint = flow.Define(goyek.Task{
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

		dockerTag := "markdownlint-cli"
		Exec(tf, dirRoot, "docker build -t "+dockerTag+" -f "+dirTools+"/markdownlint-cli.dockerfile .")
		if tf.Failed() {
			return
		}
		Exec(tf, dirRoot, "docker run --rm -v '"+curDir+":/workdir' "+dockerTag+" *.md")
	},
})
