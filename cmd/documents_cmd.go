package cmd

import (
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	errIndexNotSelected = "No index selected. Select index using 'use <index-name>'"
)

// UseIndex selects an index to use with document operations
func UseIndex() *ishell.Cmd {

	return &ishell.Cmd{
		Name: "use",
		Help: "Select index to use for subsequent document operations",
		Func: func(c *ishell.Context) {
			if context == nil {
				errorMsg(c, errNotConnected)
			} else {
				if len(c.Args) < 1 {
					errorMsg(c, "Index name not specified")
					return
				}
				s, err := context.ResolveAndValidateIndex(c.Args[0])
				if err != nil {
					errorMsg(c, err.Error())
					return
				}
				context.ActiveIndex = s
				if s != c.Args[0] {
					cprintln(c, "For alias %s selected index %s", c.Args[0], gr(s))
				} else {
					cprintln(c, "Selected index %s", s)
				}
			}
		},
	}

}

// Document is a container for document-related operations
func Document() *ishell.Cmd {
	document := &ishell.Cmd{
		Name: "document",
		Help: "Document operations",
	}

	document.AddCmd(&ishell.Cmd{
		Name: "list",
		Help: "List documents in index",
		Func: showDocs,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "properties",
		Help: "Show properties of the document",
		Func: showProperties,
	})

	return document
}

func showDocs(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	if context.ActiveIndex == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}

	docs, err := context.ListDocuments(context.ActiveIndex)

	if err != nil {
		errorMsg(c, err.Error())
	} else {
		for _, doc := range docs {
			cprintln(c, doc)
		}
	}
}

func showProperties(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	if context.ActiveIndex == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	var doc string
	if context.Version[0] > 6 {
		doc = "_doc"
	} else if len(c.Args) < 1 {
		errorMsg(c, "Please specify document name")
		return
	} else {
		doc = c.Args[0]
	}

	props, err := context.ListProperties(context.ActiveIndex, doc)

	if err != nil {
		errorMsg(c, err.Error())
	} else {
		for _, prop := range props {
			cprintln(c, "%s: %s", prop.Name, prop.Type)
		}
	}
}
