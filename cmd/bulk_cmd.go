package cmd

import (
	"fmt"
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
		Help: "Exports data into file. Usage: export [--index <index-name>] [--doc <doc-type>] <filename>",
		Func: bulkExport,
	})

	return bulk
}

func bulkExport(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}
	selector, err := parseDocumentArgs(c.Args)

	if err != nil {
		errorMsg(c, err.Error())
		return
	}

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
	go recordWriter(c, fileName, recChan, finChan)

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

func recordWriter(c *ishell.Context, fileName string, recordSupplier chan *es.BulkRecord, fin chan error) {
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		fin <- err
		return
	}
	c.ProgressBar().Start()
	c.ProgressBar().Final(bl("\nDone\n"))

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

				jsonString, err := utils.MapToJSON(rec.Content)
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
	c.ProgressBar().Suffix("      ")
	c.ProgressBar().Progress(100)
	c.ProgressBar().Stop()
}
