package cmd

import (
	"fmt"
	"shelastic/es"

	"github.com/fatih/color"
	"gopkg.in/abiosoft/ishell.v2"
)

var (
	context *es.Es

	// Commands contain list of available top-level shell commands
	Commands = []*ishell.Cmd{
		Connect(),
		Disconnect(),
		Health(),
		List(),
		Index(),
		Debug(),
	}

	cl  = color.New(color.FgBlue).SprintfFunc()
	red = color.New(color.FgRed).SprintFunc()
)

const (
	errNotConnected = "Not connected to Elasticsearch cluster"
)

// Connect performs connection to Elasticsearh cluster
func Connect() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "connect",
		Help: "Connect to ElasticSearch",
		Func: func(c *ishell.Context) {
			if context != nil {
				errorMsg(c, "Already connected to %s. Disconnect before connecting to another cluster", context.ClusterName)
				return
			}
			var host string
			if len(c.Args) < 1 {
				host = "localhost"
			} else {
				host = c.Args[0]
			}
			cprintln(c, "Connecting to %s", host)
			var err error
			var ping *es.PingResponse
			context, ping, err = es.Connect(host)
			if err == nil {
				cprintln(c, "Connected to %s (version %s)", ping.ClusterName, ping.Version)
				onConnect(context, c)
			} else {
				errorMsg(c, fmt.Sprintf("Failed to connect to %s: %s", host, err.Error()))
			}
		},
	}
}

// Disconnect disconnects from Elasticsearch cluster
func Disconnect() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "disconnect",
		Help: "Close connection to ElasticSearch",
		Func: func(c *ishell.Context) {
			cprintln(c, "Disconnected from %s", context.ClusterName)
			context = nil
			c.SetPrompt("$> ")
		},
	}
}

// Health retrieves cluster health status
func Health() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "health",
		Help: "Display cluster health information",
		Func: func(c *ishell.Context) {
			if context == nil {
				errorMsg(c, errNotConnected)
			} else {
				h, err := context.Health()
				if err != nil {
					errorMsg(c, err.Error())
				} else {
					cprintln(c, h.String())
				}
			}
		},
	}
}

// List retrieves node and indices
func List() *ishell.Cmd {
	list := &ishell.Cmd{
		Name: "list",
		Help: "List entities",
	}

	list.AddCmd(&ishell.Cmd{
		Name: "indices",
		Help: "List indices",
		Func: func(c *ishell.Context) {
			if context != nil {
				result, err := context.ListIndices()
				if err != nil {
					errorMsg(c, "Failed to retrieve list of indices: %s", err.Error())
				}
				for _, index := range result {
					cprintln(c, index.String())
				}
			} else {
				errorMsg(c, errNotConnected)
			}
		},
	})

	list.AddCmd(&ishell.Cmd{
		Name: "nodes",
		Help: "List nodes",
		Func: func(c *ishell.Context) {
			if context != nil {
				result, err := context.ListNodes()
				if err != nil {
					errorMsg(c, "Failed to retrieve list of nodes: %s", err.Error())
				} else {
					for _, index := range result {
						cprintln(c, index.String())
					}
				}
			} else {
				errorMsg(c, errNotConnected)
			}
		},
	})

	return list
}

// Debug command toggle debug mode on and off
func Debug() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "_debug",
		Help: "Toggle debug mode. Requests sent to ES cluster as well as responses are printed on the screen",
		Func: func(c *ishell.Context) {
			if context == nil {
				errorMsg(c, errNotConnected)
			} else {
				context.Debug = !context.Debug
				if context.Debug {
					cprintln(c, "Debug on")
				} else {
					cprintln(c, "Debug off")
				}
			}
		},
	}
}

func cprintln(c *ishell.Context, format string, params ...interface{}) {
	c.Println(cl(format, params...))
}

func cprintf(c *ishell.Context, format string, params ...interface{}) {
	c.Printf(cl(format, params...))
}

func errorMsg(c *ishell.Context, message string, params ...interface{}) {
	c.Println(red(fmt.Sprintf(message, params...)))
}

func onConnect(es *es.Es, c *ishell.Context) {
	health, err := es.Health()
	if err != nil {
		errorMsg(c, "Failed to retrieve Elastisearch cluster health: %s", err.Error())
		return
	}
	var colorw func(...interface{}) string
	switch health.Status {
	case "yellow":
		colorw = color.New(color.FgYellow).SprintFunc()
	case "red":
		colorw = red
	case "green":
		colorw = color.New(color.FgGreen).SprintFunc()
	}
	cprintln(c, "Status: %s", colorw(health.Status))
	c.SetPrompt(health.ClusterName + " $> ")
}
