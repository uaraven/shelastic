package main

import (
	"golastic/es"

	"gopkg.in/abiosoft/ishell.v2"
)

var (
	context *es.Es
)

func main() {
	shell := ishell.New()

	// display welcome info.
	shell.Println("Sample Interactive Shell")

	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "connect",
		Help: "Connect to ElasticSearch",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("Specify host to connect")
				c.Println("connect <host>")
			} else {
				c.Println("Connecting to", c.Args[0])
				context = es.Connect(c.Args[0])
				health := context.Health()
				c.Println("Connected to", health.ClusterName)
				c.Println("Status:", health.Status)
			}
		},
	})

	// run shell
	shell.Start()

	context = es.Connect("192.168.20.10")
	health := context.Health()
	print(health)
}
