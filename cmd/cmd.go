package cmd

import (
	"fmt"
	"shelastic/es"

	"github.com/fatih/color"
	"gopkg.in/abiosoft/ishell.v2"
)

var (
	context *es.Es

	Commands = []*ishell.Cmd{
		Connect(),
		List(),
	}

	red = color.New(color.FgRed).SprintFunc()
)

const (
	errNotConnected = "Not connected to Elasticsearch cluster"
)

func Connect() *ishell.Cmd {
	return &ishell.Cmd{
		Name: "connect",
		Help: "Connect to ElasticSearch",
		Func: func(c *ishell.Context) {
			var host string
			if len(c.Args) < 1 {
				host = "localhost"
			} else {
				host = c.Args[0]
			}
			c.Println("Connecting to", host)
			var err error
			var ping *es.PingResponse
			context, ping, err = es.Connect(host)
			if err == nil {
				c.Println(fmt.Sprintf("Connected to %s (version %s)", ping.ClusterName, ping.Version))
				onConnect(context, c)
			} else {
				errorMsg(c, fmt.Sprintf("Failed to connect to %s: %s", host, err.Error()))
			}
		},
	}
}

func List() *ishell.Cmd {
	list := &ishell.Cmd{
		Name: "list",
		Help: "List entities",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("Specify what to list")
				c.Println("list indices|nodes")
			}
		},
	}

	list.AddCmd(&ishell.Cmd{
		Name: "indices",
		Help: "List indices",
		Func: func(c *ishell.Context) {
			if context != nil {
				result, err := context.ListIndices()
				if err != nil {
					errorMsg(c, "Failed to retrieve list of indices: "+err.Error())
				}
				for _, index := range result {
					c.Println(index)
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
					errorMsg(c, "Failed to retrieve list of nodes: "+err.Error())
				} else {
					for _, index := range result {
						c.Println(index.String())
					}
				}
			} else {
				errorMsg(c, errNotConnected)
			}
		},
	})

	return list
}

func errorMsg(c *ishell.Context, message string) {
	c.Println(red(message))
}

func onConnect(es *es.Es, c *ishell.Context) {
	health, err := es.Health()
	if err != nil {
		errorMsg(c, "Failed to retrieve Elastisearch cluster health: "+err.Error())
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
	c.Println("Status:", colorw(health.Status))
	c.SetPrompt(health.ClusterName + " $> ")
}
