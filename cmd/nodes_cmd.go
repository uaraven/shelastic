package cmd

import (
	ishell "gopkg.in/abiosoft/ishell.v2"
)

// Nodes wraps all nodes functions
func Nodes() *ishell.Cmd {
	snapshot := &ishell.Cmd{
		Name: "node",
		Help: "Node operations",
	}

	snapshot.AddCmd(&ishell.Cmd{
		Name: "stats",
		Help: "Get node statistics",
		Func: getNodeStats,
	})

	return snapshot
}

func getNodeStats(c *ishell.Context) {
	if context != nil {
		nodeStats, err := context.GetNodeStats(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		cprintln(c, nodeStats.String(cl))
	} else {
		errorMsg(c, errNotConnected)
	}
}
