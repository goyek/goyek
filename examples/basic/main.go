package main

import (
	"github.com/goyek/goyek"
)

func main() {
	flow := &goyek.Flow{}

	flow.Register(goyek.Task{
		Name:  "hello",
		Usage: "demonstration",
		Action: func(tf *goyek.TF) {
			tf.Log("Hello world!")
		},
	})

	flow.Main()
}
