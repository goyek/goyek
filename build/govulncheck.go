package main

import "github.com/goyek/goyek/v3"

var govulncheck = goyek.Define(goyek.Task{
	Name:  "govulncheck",
	Usage: "scan for known vulnerabilities",
	Action: func(a *goyek.A) {
		if !Exec(a, dirBuild, "go", "install", "golang.org/x/vuln/cmd/govulncheck") {
			return
		}
		Exec(a, dirRoot, "govulncheck", "./...")
	},
})
