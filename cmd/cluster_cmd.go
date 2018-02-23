package cmd

import (
	"shelastic/utils"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

// Cluster is a parent for cluster-related operations
func Cluster() *ishell.Cmd {
	cluster := &ishell.Cmd{
		Name: "cluster",
		Help: "Cluster operations",
	}

	cluster.AddCmd(&ishell.Cmd{
		Name: "health",
		Help: "Display cluster health status",
		Func: health,
	})

	cluster.AddCmd(&ishell.Cmd{
		Name: "settings",
		Help: "Display cluster settings",
		Func: clusterSettings,
	})

	return cluster
}

func health(c *ishell.Context) {
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
}

func clusterSettings(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
	} else {
		settings, err := context.GetSettings()
		if err != nil {
			errorMsg(c, err.Error())
		}
		yaml, err := utils.MapToYaml(settings)
		if err != nil {
			errorMsg(c, err.Error())
		}
		cprintln(c, yaml)
	}
}
