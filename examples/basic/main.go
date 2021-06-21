package main

import (
	"github.com/goyek/goyek"
)

func main() {
	flow := &goyek.Flow{}

	flow.Register(goyek.Task{
		Name:  "hello",
		Usage: "demonstration",
		Action: func(p *goyek.Progress) {
			p.Log("Hello world!")
		},
	})

	flow.Main()
}
