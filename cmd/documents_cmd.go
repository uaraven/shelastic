package cmd

import (
	"encoding/json"

	ishell "gopkg.in/abiosoft/ishell.v2"
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
						cprintlist(c, "Using index ", cy(context.ActiveIndex))
					} else {
						cprintln(c, "No index is in use")
					}
					return
				}
				if c.Args[0] == "--" {
					if context.ActiveIndex != "" {
						cprintlist(c, "Index ", cy(context.ActiveIndex), " is no longer in use")
						context.ActiveIndex = ""
					} else {
						cprintln(c, "No index is in use")
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
					cprintlist(c, "For alias ", cyb(c.Args[0]), " selected index ", cy(s))
				} else {
					cprintlist(c, "Selected index ", cy(s))
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
		Help: "List documents in index. Usage: list [--index <index-name>]",
		Func: showDocs,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "properties",
		Help: "Show properties of the document. Usage: properties [--index <index-name>] [--doc] <type>",
		Func: showProperties,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "Retrieves document by its id. Usage: get [--index <index-name>] --doc <type> <id>",
		Func: getDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "put",
		Help: "Inserts/updates document. Usage: put [--index <index-name>] --doc <type> <id>",
		Func: putDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "delete",
		Help: "Deletes document by its id. Usage: delete [--index <index-name>] --doc <type> <id>",
		Func: deleteDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "Peforms simple search. Usage: search [--index <index-name>] [--doc <types>] <search string>",
		Func: searchDocument,
	})

	document.AddCmd(&ishell.Cmd{
		Name: "query",
		Help: "Peforms search using query DSL. Usage: query [--index <index-name>]",
		Func: queryDocument,
	})

	return document
}

func showDocs(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	if selector.Index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}

	docs, err := context.ListDocuments(selector.Index)

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
	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	if selector.Index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	var doc string
	if context.Version[0] > 6 {
		doc = "_doc"
	} else if selector.Document != "" {
		doc = selector.Document
	} else if len(selector.Args) > 0 {
		doc = selector.Args[0]
	}

	if doc == "" {
		errorMsg(c, "Please specify document name using --doc <document-name> parameter")
		return
	}

	props, err := context.ListProperties(selector.Index, doc)

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
	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	if selector.Index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.Args) == 0 || selector.Document == "" {
		errorMsg(c, "Not enough parameters. Usage: get [--index <index-name>] --doc <doc-type> <id>")
		return
	}
	doc, err := context.GetDocument(selector.Index, selector.Document, selector.Args[0])
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
	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	if selector.Index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.Args) == 0 || selector.Document == "" {
		errorMsg(c, "Not enough parameters. Usage: put [--index <index-name>] --doc <doc-type> <id>")
		return
	}
	cprintln(c, "Enter document body, ending with ';':")
	c.SetPrompt(">>> ")
	json := c.ReadMultiLines(";")
	json = json[:len(json)-1]
	restorePrompt(context, c)
	response, err := context.PutDocument(selector.Index, selector.Document, selector.Args[0], json)
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
	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	if selector.Index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.Args) == 0 || selector.Document == "" {
		errorMsg(c, "Not enough parameters. Usage: delete [--index <index-name>] --doc <doc-type> <id>")
		return
	}
	err = context.DeleteDocument(selector.Index, selector.Document, selector.Args[0])
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
	selector, err := parseDocumentArgs(c.Args)
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	if selector.Index == "" {
		errorMsg(c, errIndexNotSelected)
		return
	}
	if len(selector.Args) == 0 {
		errorMsg(c, "Not enough parameters. Usage: search [--index <index-name>] [--doc <doc-types>] <search query>")
		return
	}
	sr, err := context.Search(selector.Index, selector.Document, selector.Args[0])
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, "Total hits: %d\n", sr.Total)
	for _, hit := range sr.Hits {
		cprintln(c, hit)
	}
}

func queryDocument(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	selector, err := parseDocumentArgs(c.Args)

	if err != nil {
		errorMsg(c, err.Error())
		return
	}

	cprintln(c, "Enter query, ending with ';'")
	c.SetPrompt(">>> ")
	q := c.ReadMultiLines(";")
	q = q[:len(q)-1]
	restorePrompt(context, c)

	var body map[string]interface{}

	if err := json.Unmarshal([]byte(q), &body); err != nil {
		errorMsg(c, "Invalid query JSON: "+err.Error())
		return
	}

	body["size"] = 20
	bytes, err := json.Marshal(body)

	if err != nil {
		errorMsg(c, "Failed to repack query JSON: "+err.Error())
		return
	}

	sr, err := context.Query(selector.Index, string(bytes))
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	cprintln(c, "Total hits: %d\n", sr.Total)
	for _, hit := range sr.Hits {
		cprintln(c, hit)
	}
}
