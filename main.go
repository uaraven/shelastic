package main

import (
	"shelastic/cmd"

	"gopkg.in/abiosoft/ishell.v2"
)

func main() {
	shell := ishell.New()

	shell.SetHomeHistoryPath(".shelastic_history")
	shell.SetPrompt("$> ")
	shell.ShowPrompt(true)

	// display welcome info.
	shell.Println("Elasticsearch shell")

	for _, c := range cmd.Commands {
		shell.AddCmd(c)
	}

	// run shell
	shell.Start()
}
