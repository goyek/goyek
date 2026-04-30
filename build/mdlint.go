package main

import (
	"os"
	"os/exec"

	"github.com/goyek/goyek/v3"
)

var mdlint = goyek.Define(goyek.Task{
	Name:  "mdlint",
	Usage: "markdownlint-cli (uses docker)",
	Action: func(a *goyek.A) {
		if _, err := exec.LookPath("docker"); err != nil {
			a.Skip(err)
		}
		curDir, err := os.Getwd()
		if err != nil {
			a.Fatal(err)
		}
		mdFiles := find(a, ".md")
		if len(mdFiles) == 0 {
			a.Skip("no .md files")
		}
		dockerImage := "ghcr.io/igorshubovych/markdownlint-cli:v0.41.0"
		args := []string{"run", "--rm", "-v", curDir + ":/workdir", dockerImage}
		args = append(args, mdFiles...)
		ExecArgs(a, dirRoot, "docker", args...)
	},
})
