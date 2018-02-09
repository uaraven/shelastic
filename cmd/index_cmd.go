package cmd

import (
	ishell "gopkg.in/abiosoft/ishell.v2"
)

func Index() *ishell.Cmd {
	index := &ishell.Cmd{
		Name: "index",
		Help: "Index operations",
	}

	view := &ishell.Cmd{
		Name: "view",
		Help: "View index data",
	}

	view.AddCmd(&ishell.Cmd{
		Name: "mapping",
		Help: "View index mapping. Usage: index view mapping <index-name> [<doc> [<property>]]",
		Func: viewIndexMapping,
	})

	view.AddCmd(&ishell.Cmd{
		Name: "settings",
		Help: "View index settings. Usage: index view settings <index-name>",
		Func: viewIndexSettings,
	})

	index.AddCmd(view)
	index.AddCmd(&ishell.Cmd{
		Name: "flush",
		Help: "Flushes index data to storage and clears transaction log",
		Func: flush,
	})

	return index
}

func flush(c *ishell.Context) {
	if context != nil {
		var index string
		if len(c.Args) == 0 {
			index = ""
		} else {
			index = c.Args[0]
		}
		err := context.Flush(index, false, false)
		if err != nil {
			errorMsg(c, err.Error())
		} else {
			cprintln(c, "Ok")
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func viewIndexMapping(c *ishell.Context) {
	if context != nil {
		indexName := ""
		docType := ""
		property := ""

		if len(c.Args) >= 1 {
			indexName = c.Args[0]
		}
		if len(c.Args) >= 2 {
			docType = c.Args[1]
		}
		if len(c.Args) >= 3 {
			property = c.Args[2]
		}
		if indexName == "" {
			errorMsg(c, "Index name not specified")
			return
		}

		result, err := context.IndexViewMapping(indexName, docType, property)
		if err != nil {
			errorMsg(c, err.Error())
		}
		cprintln(c, result)

	} else {
		errorMsg(c, errNotConnected)
	}
}

func viewIndexSettings(c *ishell.Context) {
	if context != nil {

		if len(c.Args) < 1 {
			errorMsg(c, "Index name not specified")
			return
		}
		indexName := c.Args[0]

		result, err := context.IndexViewSettings(indexName)
		if err != nil {
			errorMsg(c, err.Error())
		}
		cprintln(c, "Number of shards: %d\nNumber of replicas: %d", result.NumberOfShards, result.NumberOfReplicas)

	} else {
		errorMsg(c, errNotConnected)
	}
}
