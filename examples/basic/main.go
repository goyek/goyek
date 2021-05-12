package main

import (
	"fmt"
	"os"

	"github.com/goyek/goyek"
)

func main() {
	if err := os.Chdir(".."); err != nil {
		fmt.Println(err)
		os.Exit(goyek.CodeInvalidArgs)
	}

	flow := &goyek.Taskflow{}

	flow.Register(goyek.Task{
		Name:  "hello",
		Usage: "demonstration",
		Command: func(tf *goyek.TF) {
			tf.Log("Hello world!")
		},
	})

	flow.Main()
}
