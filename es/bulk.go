package es

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	scrollLength = "2m"
	size         = 20
)

// BulkRecord contains record returned from ES and progress counter in percents
type BulkRecord struct {
	ID       string
	Index    string
	Document string
	Content  map[string]interface{}
	Progress int
}

// BulkExport performs ES scroll search and exports records into "output" channel. Channel is closed after all records are exported
func (e Es) BulkExport(index string, doc string, query string, output chan *BulkRecord, ctlChan chan error) {
	if index != "" {
		index = "/" + index
	}
	var body map[string]interface{}

	if err := json.Unmarshal([]byte(query), &body); err != nil {
		ctlChan <- fmt.Errorf("Invalid query JSON: %s", err.Error())
		return
	}

	body["size"] = size
	if (e.Version[0] == 2 && e.Version[1] >= 1) || e.Version[0] >= 5 {
		body["sort"] = []string{"_doc"}
	}

	bytes, err := json.Marshal(body)

	if err != nil {
		ctlChan <- fmt.Errorf("Failed to repack query JSON: %s", err.Error())
		return
	}

	resp, err := e.getJSONWithBody(fmt.Sprintf("%s%s/_search?scroll=%s", index, doc, scrollLength), string(bytes))
	if err != nil {
		ctlChan <- err
		return
	}
	err = checkError(resp)
	if err != nil {
		ctlChan <- err
		return
	}
	var scrollID string

	if e.Version[0] < 2 {
		// versions 1.x do not return records in first scroll request, so we skip to next one immediately
		scrollID = resp["_scroll_id"].(string)
		resp, err = e.getJSON(fmt.Sprintf("/_search/scroll?scroll=%s&scroll_id=%s", scrollLength, scrollID))
		if err != nil {
			ctlChan <- err
			return
		}
	}
	count := 0

	for {
		scrollID, ok := resp["_scroll_id"].(string)
		if !ok {
			ctlChan <- fmt.Errorf("Unexpected response: Response does not contain _scroll_id")
			return
		}

		hits, ok1 := resp["hits"].(map[string]interface{})
		totalf, ok2 := hits["total"].(float64)
		if !(ok1 || ok2) {
			ctlChan <- fmt.Errorf("Unexpected response: no hits or total")
			return
		}
		total := int(totalf)
		records := hits["hits"].([]interface{})
		if !ok {
			ctlChan <- fmt.Errorf("Unexpected response: No hits")
			return
		}

		if e.Debug {
			fmt.Printf("Bulk supplier: Total: %d, Records: %d\n", total, len(records))
		}

		if len(records) == 0 {
			break
		}

		for _, record := range records {
			count++
			res := &BulkRecord{
				Content:  record.(map[string]interface{}),
				Progress: (count * 100) / total,
			}

			// check if we are still in a game?
			select {
			case errValue, isOk := <-ctlChan:
				if errValue != nil || !isOk {
					if e.Debug {
						fmt.Println("Bulk supplier: Closed channel")
					}
					return
				}
			default:
				// nothing on error channel, we may continue
			}

			if e.Debug {
				fmt.Println("Bulk supplier: Sending record")
			}

			output <- res
		}

		if e.Version[0] >= 2 {
			resp, err = e.getJSONWithBody("/_search/scroll", fmt.Sprintf("{\"scroll\":\"%s\",\"scroll_id\":\"%s\"}", scrollLength, scrollID))
		} else {
			resp, err = e.getJSON(fmt.Sprintf("/_search/scroll?scroll=%s&scroll_id=%s", scrollLength, scrollID))
		}
		if err != nil {
			ctlChan <- err
			return
		}
	}

	if e.Debug {
		fmt.Println("Bulk supplier: finished")
	}

	ctlChan <- nil
}

//BulkImport reads data from file, converts to ES bulk format and executes bulk insert
func (e Es) BulkImport(indexName string, documentName string, idfld string, data string) error {
	scnr := bufio.NewScanner(strings.NewReader(data))
	wrtr := new(bytes.Buffer)

	for scnr.Scan() {
		line := scnr.Text()
		recJSON := make(map[string]interface{})
		err := json.Unmarshal([]byte(line), &recJSON)
		if err != nil {
			return err
		}
		var idstr string
		if idfld != "" {
			id, ok := recJSON[idfld].(string)
			if !ok {
				return fmt.Errorf("No field '%s' in record", idfld)
			}
			idstr = fmt.Sprintf(", \"_id\": %s", id)
		} else {
			idstr = ""
		}
		wrtr.WriteString(fmt.Sprintf("{\"index\":{\"_index\" : \"%s\", \"_doc\":\"%s\" %s}}\n", indexName, documentName, idstr))
		wrtr.WriteString(line)
	}

	bulkBody := wrtr.String()

	resp, err := e.postJSON("/_bulk", bulkBody)

	if err != nil {
		return err
	}

	err = checkError(resp)

	return err
}
