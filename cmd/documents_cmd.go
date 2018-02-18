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
					if context.ActiveIndex != "" {
						cprintln(c, "Using index %s", gr(context.ActiveIndex))
					} else {
						errorMsg(c, "Index name not specified")
					}
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
		Help: "Show properties of the document. Usage: properties <type>",
		Func: showProperties,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "Retrieves document by its id. Usage: get <type> <id>",
		Func: getDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "delete",
		Help: "Deletes document by its id. Usage: delete <type> <id>",
		Func: deleteDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "Peforms simple search. Usage: search [<types>] <search string>",
		Func: searchDocument,
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

func getDocument(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	if context.ActiveIndex == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(c.Args) < 2 {
		errorMsg(c, "Not enough parameters. Usage: get <doc-type> <id>")
		return
	}
	doc, err := context.GetDocument(context.ActiveIndex, c.Args[0], c.Args[1])
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, doc)
}

func deleteDocument(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	if context.ActiveIndex == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(c.Args) < 2 {
		errorMsg(c, "Not enough parameters. Usage: delete <doc-type> <id>")
		return
	}
	err := context.DeleteDocument(context.ActiveIndex, c.Args[0], c.Args[1])
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, "Ok")
}

func searchDocument(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	if context.ActiveIndex == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(c.Args) < 1 {
		errorMsg(c, "Not enough parameters. Usage: search [<doc-types>] <search query>")
		return
	}
	var docs string
	var query string
	if len(c.Args) == 1 {
		docs = ""
		query = c.Args[0]
	} else {
		docs = c.Args[0]
		query = c.Args[1]
	}
	sr, err := context.Search(context.ActiveIndex, docs, query)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, "Total hits: %d\n", sr.Total)
	for _, hit := range sr.Hits {
		cprintln(c, hit)
	}
}
