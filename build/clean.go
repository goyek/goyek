package main

import (
	"os"

	"github.com/goyek/goyek/v2"
)

var _ = goyek.Define(goyek.Task{
	Name:  "clean",
	Usage: "remove files created during build pipeline",
	Action: func(a *goyek.A) {
		remove(a, "coverage.out")
		remove(a, "coverage.html")
	},
})

func remove(a *goyek.A, path string) {
	a.Helper()
	if _, err := os.Stat(path); err != nil {
		return
	}
	a.Log("Remove: " + path)
	if err := os.RemoveAll(path); err != nil {
		a.Error(err)
	}
}
