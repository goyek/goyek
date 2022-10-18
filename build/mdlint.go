package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

		var mdFiles []string
		err = filepath.WalkDir(dirRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(d.Name()) == ".md" {
				mdFiles = append(mdFiles, path)
			}
			return nil
		})
		if err != nil {
			tf.Fatal(err)
		}

		if len(mdFiles) > 0 {
			dockerImage := "ghcr.io/igorshubovych/markdownlint-cli:v0.32.2"
			Exec(tf, dirRoot, "docker run --rm -v '"+curDir+":/workdir' "+dockerImage+" "+strings.Join(mdFiles, " "))
		}
	},
})
