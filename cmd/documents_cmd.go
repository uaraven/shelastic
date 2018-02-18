package cmd

import (
	"fmt"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	errIndexNotSelected = "No index selected. Select index using 'use <index-name>'"
)

type documentSelector struct {
	index    string
	document string
	rest     []string
}

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
		Help: "Show properties of the document. Usage: properties [<index>] <type>",
		Func: showProperties,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "Retrieves document by its id. Usage: get [<index>] <type> <id>",
		Func: getDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "put",
		Help: "Inserts/updates document. Usage: put[<index>]  <type> <id>",
		Func: putDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "delete",
		Help: "Deletes document by its id. Usage: delete [<index>]  <type> <id>",
		Func: deleteDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "Peforms simple search. Usage: search [<index>] [<types>] <search string>",
		Func: searchDocument,
	})

	return document
}

func showDocs(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	selector := parseArguments(c.Args, 0)
	if selector.index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}

	docs, err := context.ListDocuments(selector.index)

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
	selector := parseArguments(c.Args, 1)
	if selector.index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	var doc string
	if context.Version[0] > 6 {
		doc = "_doc"
	} else if selector.document == "" {
		errorMsg(c, "Please specify document name")
		return
	} else {
		doc = selector.document
	}

	props, err := context.ListProperties(selector.index, doc)

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
	selector := parseArguments(c.Args, 2)
	if selector.index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.rest) == 0 || selector.document == "" {
		errorMsg(c, "Not enough parameters. Usage: get [index] <doc-type> <id>")
		return
	}
	doc, err := context.GetDocument(selector.index, selector.document, selector.rest[0])
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, doc)
}

func putDocument(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	selector := parseArguments(c.Args, 2)
	if selector.index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.rest) == 0 || selector.document == "" {
		errorMsg(c, "Not enough parameters. Usage: get [index] <doc-type> <id>")
		return
	}
	c.SetPrompt(">>> ")
	json := c.ReadMultiLines(";")
	json = json[:len(json)-1]
	fmt.Println(json)
	restorePrompt(context, c)
	response, err := context.PutDocument(selector.index, selector.document, selector.rest[0], json)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, response)
}

func deleteDocument(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	selector := parseArguments(c.Args, 2)
	if selector.index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.rest) == 0 || selector.document == "" {
		errorMsg(c, "Not enough parameters. Usage: delete [index] <doc-type> <id>")
		return
	}
	err := context.DeleteDocument(selector.index, selector.document, selector.rest[0])
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
	selector := parseArguments(c.Args, 2)
	if selector.index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if selector.index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.rest) == 0 {
		errorMsg(c, "Not enough parameters. Usage: search [index] [<doc-types>] <search query>")
		return
	}
	sr, err := context.Search(selector.index, selector.document, selector.rest[0])
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, "Total hits: %d\n", sr.Total)
	for _, hit := range sr.Hits {
		cprintln(c, hit)
	}
}

// parseArguments parses list of arguments into index name, document type and the rest or args
// If 3 or more parameters passed, first is treated as index name, second is document
// If less than 3 parameters passed, then first is treated as document name and index name is retrieved from context.ActiveIndex
func parseArguments(args []string, expectedArgs int) *documentSelector {
	var index string
	var doc string
	if len(args) > expectedArgs {
		index = args[0]
		args = args[1:]
	} else {
		index = context.ActiveIndex
	}
	if len(args) == expectedArgs && expectedArgs > 0 {
		doc = args[0]
		args = args[1:]
	} else {
		doc = ""
	}
	return &documentSelector{
		index:    index,
		document: doc,
		rest:     args,
	}
}
