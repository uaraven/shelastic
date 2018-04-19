package main

import (
	"runtime"
	"shelastic/cmd"

	flags "github.com/jessevdk/go-flags"
	"gopkg.in/abiosoft/ishell.v2"
)

const (
	// Version of shelastic
	Version = "0.3.1"
)

func main() {
	shell := ishell.New()

	shell.SetHomeHistoryPath(".shelastic_history")
	shell.SetPrompt("$> ")
	shell.ShowPrompt(true)

	// display welcome info.
	shell.Println("Shelastic [Elasticsearch shell]", "v"+Version)

	for _, c := range cmd.Commands {
		shell.AddCmd(c)
	}

	settings := &cmd.Settings{}

	_, err := flags.Parse(settings)

	if runtime.GOOS == "windows" {
		settings.NoColor = true
	}

	if err == nil {
		cmd.Initialize(settings)
		shell.Start()
	}

}
