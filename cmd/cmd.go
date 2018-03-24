package cmd

import (
	"fmt"
	"reflect"
	"shelastic/es"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/abiosoft/ishell.v2"
)

var (
	context *es.Es

	// Commands contain list of available top-level shell commands
	Commands = []*ishell.Cmd{
		Connect(),
		Disconnect(),
		Cluster(),
		List(),
		Index(),
		Snapshot(),
		Nodes(),
		Debug(),
		UseIndex(),
		Document(),
		Bulk(),
	}

	bl   = color.New(color.FgBlue).SprintfFunc()
	hbl  = color.New(color.FgHiBlue).SprintfFunc()
	red  = color.New(color.FgRed).SprintFunc()
	undr = color.New(color.Underline).SprintfFunc()
	cy   = color.New(color.FgCyan).SprintfFunc()
	cyb  = color.New(color.FgHiCyan).SprintfFunc()
	gre  = color.New(color.FgGreen).SprintFunc()
	yel  = color.New(color.FgYellow).SprintFunc()
)

const (
	errNotConnected = "Not connected to Elasticsearch cluster"
)

// Settings contains shelastic shell settings configured through command line parameters
type Settings struct {
	NoColor bool `short:"n" long:"no-color" description:"Do not use colors in terminal"`
}

// Initialize sets up internal state of the shell. Must be called before starting the shell
func Initialize(settings *Settings) {
	color.NoColor = settings.NoColor
}

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
				cprintlist(c, "Connected to ", cyb(ping.ClusterName), " (version ", ping.Version, ")")
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
					var aliases = make([]string, len(index.Aliases))
					for i, alias := range index.Aliases {
						aliases[i] = alias.Name
					}
					aliasesstr := strings.Join(aliases, ", ")
					cprintlist(c, cyb(index.Name), " [",
						fmt.Sprintf("docs: %d, bytes: %d, aliases: %s", index.DocumentCount, index.Size, hbl(aliasesstr)),
						"]")

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
					for _, node := range result {
						cprintlist(c, cyb(node.Name), " @ ", hbl(node.Host), " ["+node.IP+"]")
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
	c.Println(bl(format, params...))
}

func cprintf(c *ishell.Context, format string, params ...interface{}) {
	c.Printf(bl(format, params...))
}

// cprintlist prints list of parameters on a line. If parameter is a function it is printed as is, otherwise it is wrapped in default color
// After all items are printed, new line is printed
func cprintlist(c *ishell.Context, params ...interface{}) {
	for _, item := range params {
		if reflect.TypeOf(item).Kind() == reflect.Func {
			c.Print(item)
		} else {
			c.Print(bl("%v", item))
		}
	}
	c.Println()
}

// cprintl prints list of parameters on a line. If parameter is a function it is printed as is, otherwise it is wrapped in default color
// No new line after all is printed
func cprintl(c *ishell.Context, params ...interface{}) {
	for _, item := range params {
		if reflect.TypeOf(item).Kind() == reflect.Func {
			c.Print(item)
		} else {
			c.Print(bl("%v", item))
		}
	}
	c.Print(" ")
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
		colorw = yel
	case "red":
		colorw = red
	case "green":
		colorw = gre
	}
	cprintln(c, "Status: %s", colorw(health.Status))
	c.SetPrompt(health.ClusterName + " $> ")
}

func restorePrompt(c *ishell.Context) {
	if context.ActiveIndex != "" {
		c.SetPrompt(fmt.Sprintf("%s.%s $> ", context.ClusterName, context.ActiveIndex))
	} else {
		c.SetPrompt(fmt.Sprintf("%s $> ", context.ClusterName))
	}
}

func dangerousPrompt(c *ishell.Context, text string) bool {
	cprintlist(c, text+" ", hbl("Do you want to proceed (yes/No)?"))
	c.SetPrompt("> ")
	defer restorePrompt(c)
	confirmation := strings.ToLower(c.ReadLine())
	if confirmation == "yes" {
		return true
	}
	if confirmation != "no" {
		cprintlist(c, "You need to type ", hbl("yes"), " to confirm")
	}
	return false
}
