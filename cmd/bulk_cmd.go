package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"shelastic/es"
	"shelastic/utils"
	"strings"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

// Bulk is a parent for bulk-related operations
func Bulk() *ishell.Cmd {
	bulk := &ishell.Cmd{
		Name: "bulk",
		Help: "Bulk data operations",
	}

	bulk.AddCmd(&ishell.Cmd{
		Name: "export",
		Help: "Exports data into file. Usage: export [--index <index-name>] [--doc <doc-type>] [--source] <filename>",
		Func: bulkExport,
	})

	bulk.AddCmd(&ishell.Cmd{
		Name: "import",
		Help: "Imports data from file. Usage: import [--index <index-name>] --doc <doc-type> [--id-field <id-field>] <filename>",
		Func: bulkImport,
	})

	return bulk
}

func bulkImport(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	type bulkArgs struct {
		documentSelectorData
		IDField string `long:"id-field" description:"Name of the field of the object containing the id" value-name:"ID"`
	}

	slctr, err := parseDocumentArgsCustom(c.Args, &bulkArgs{})
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	selector := slctr.(*bulkArgs)

	if selector.Index == "" {
		errorMsg(c, "Index not specified")
		return
	}
	if selector.Document == "" {
		errorMsg(c, "Document not specified")
		return
	}
	if len(selector.Args) == 0 {
		errorMsg(c, "Data file name is not specified")
		return
	}
	cprintln(c, "Reading and inserting data")

	data, err := ioutil.ReadFile(selector.Args[0])
	if err != nil {
		errorMsg(c, "Failed to read from "+selector.Args[0])
	}

	err = context.BulkImport(selector.Index, selector.Document, selector.IDField, string(data))

	if err != nil {
		errorMsg(c, "Failed to bulk insert data from "+selector.Args[0])
	} else {
		cprintln(c, "Done")
	}

}

func bulkExport(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	type bulkArgs struct {
		documentSelectorData
		Source bool `long:"source"  description:"Export only '_source' attribute"`
	}

	slctr, err := parseDocumentArgsCustom(c.Args, &bulkArgs{})

	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	selector := slctr.(*bulkArgs)
	if len(selector.Args) == 0 {
		errorMsg(c, "Not enough parameters. Usage: export <filename>")
		return
	}

	fileName := selector.Args[0]

	cprintln(c, "Enter query, ending with ';'")
	c.SetPrompt(">>> ")
	defer restorePrompt(c)
	q := c.ReadMultiLines(";")
	if len(q) > 0 {
		q = q[:len(q)-1]
	} else {
		errorMsg(c, "Invalid query")
		return
	}

	if len(q) == 0 {
		q = "{\"query\": {\"match_all\":{}}}"
		cprintln(c, "Using match all query")
	}

	recChan := make(chan *es.BulkRecord, 50)
	errChan := make(chan error)
	finChan := make(chan error)
	defer close(recChan)
	defer close(finChan)
	defer close(errChan)

	go context.BulkExport(selector.Index, selector.Document, q, recChan, errChan)
	go recordWriter(c, fileName, selector.Source, recChan, finChan)

	select {
	case err = <-errChan:
		if err != nil {
			errorMsg(c, "ES read error: "+err.Error())
		}
		finChan <- fmt.Errorf("stop")
	case err = <-finChan:
		if err != nil {
			errorMsg(c, "Write error: "+err.Error())
		}
		errChan <- fmt.Errorf("stop")
	}
}

func recordWriter(c *ishell.Context, fileName string, source bool, recordSupplier chan *es.BulkRecord, fin chan error) {
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		fin <- err
		return
	}
	c.ProgressBar().Start()

	defer stopProgress(c)

	for {
		select {
		case err, stillWorking := <-fin:
			if err != nil || !stillWorking {
				if context.Debug {
					fmt.Println("Bulk writer: Stopping writer")
				}
				return
			}
		case rec := <-recordSupplier:
			{
				if context.Debug {
					fmt.Println("Bulk writer: Received record")
				}

				c.ProgressBar().Suffix(fmt.Sprint(" ", rec.Progress, "%"))
				c.ProgressBar().Progress(rec.Progress)

				var body map[string]interface{}
				if source {
					var ok bool
					body, ok = rec.Content["_source"].(map[string]interface{})
					if !ok {
						fin <- fmt.Errorf("Search results does not contain '_source' field")
						break
					}
				} else {
					body = rec.Content
				}

				jsonString, err := utils.MapToJSON(body)
				if err != nil {
					fin <- err
					break
				}
				singleJSONString := strings.Replace(jsonString, "\n", "", -1)
				f.WriteString(singleJSONString + "\n")
			}
		}

	}
}

func stopProgress(c *ishell.Context) {
	c.ProgressBar().Suffix("  Done\n")
	c.ProgressBar().Progress(100)
	c.ProgressBar().Stop()
}
