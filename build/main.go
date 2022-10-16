// Build is the build pipeline for this repository.
package main

import (
	"os"

	"github.com/goyek/goyek/v2"
)

const (
	rootDir  = "."
	buildDir = "build"
	toolsDir = "tools"
)

var flow = &goyek.Flow{}

func main() {
	flow.SetDefault(all)
	flow.Main(os.Args[1:])
}
