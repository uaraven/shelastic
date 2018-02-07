package cmd

import (
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
			if len(c.Args) < 1 {
				c.Println("Specify host to connect")
				c.Println("connect <host>")
			} else {
				c.Println("Connecting to", c.Args[0])
				var err error
				context, err = es.Connect(c.Args[0])
				if err == nil {
					onConnect(context, c)
				} else {
					errorMsg(c, "Failed to connect to "+c.Args[0]+": "+err.Error())
				}
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
					errorMsg(c, "Failed to retrieve list of indices:"+err.Error())
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
					errorMsg(c, "Failed to retrieve list of nodes:"+err.Error())
				}
				for _, index := range result {
					c.Println(index)
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
	c.Println("Connected to", health.ClusterName)
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
