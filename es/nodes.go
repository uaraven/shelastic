package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// NodeJvmStats contains JVM statistics for node
type NodeJvmStats struct {
	Uptime  int64            `json:"uptime_in_millis"`
	Memory  *NodeMemoryStats `json:"mem"`
	Threads *NodeThreadStats `json:"threads"`
}

// NodeMemoryStats contains JVM memory statistics for node
type NodeMemoryStats struct {
	HeapUsed         int64 `json:"heap_used_in_bytes"`
	HeapCommitted    int64 `json:"heap_committed_in_bytes"`
	HeapMax          int64 `json:"heap_max_in_bytes"`
	NonHeapUsed      int64 `json:"non_heap_used_in_bytes"`
	NonHeapCommitted int64 `json:"non_heap_committed_in_bytes"`
}

// NodeThreadStats contains JVM thread statistics for node
type NodeThreadStats struct {
	Count     int64 `json:"count"`
	PeakCount int64 `json:"peak_count"`
}

// NodeIndicesStats contains index statistics for node
type NodeIndicesStats struct {
	Docs struct {
		Count   int64 `json:"count"`
		Deleted int64 `json:"deleted"`
	} `json:"docs"`
	Store struct {
		Size         int64 `json:"size_in_bytes"`
		ThrottleTime int64 `json:"throttle_time_in_millis"`
	} `json:"store"`
}

// FsStats contains filesystem statistics
type FsStats struct {
	Total struct {
		Total     int64 `json:"total_in_bytes"`
		Free      int64 `json:"free_in_bytes"`
		Available int64 `json:"available_in_bytes"`
	} `json:"total"`
	Data []FsNodeStats `json:"data"`
}

// FsNodeStats contains filesystem statistics per node
type FsNodeStats struct {
	Path      string `json:"path"`
	Mount     string `json:"mount"`
	Type      string `json:"type"`
	Total     int64  `json:"total_in_bytes"`
	Free      int64  `json:"free_in_bytes"`
	Available int64  `json:"available_in_bytes"`
}

// NodeStats holds Statistics for a single node
type NodeStats struct {
	Name             string           `json:"name"`
	TransportAddress string           `json:"transport_address"`
	Host             string           `json:"host"`
	IP               []string         `json:"ip"`
	Indices          NodeIndicesStats `json:"indices"`
	JVM              NodeJvmStats     `json:"jvm"`
	FS               FsStats          `json:"fs"`
}

// NodesStats is a main container holding statistics for all nodes
type NodesStats struct {
	Nodes map[string]NodeStats `json:"nodes"`
}

// GetNodeStats Retrieves node statistics from Elasticsearch
func (e Es) GetNodeStats(nodes []string) (*NodesStats, error) {
	nodestr := strings.Join(nodes, ",")
	var url string
	if len(nodes) > 0 {
		url = fmt.Sprintf("_nodes/%s/stats/", nodestr)
	} else {
		url = "_nodes/stats/"
	}

	body, err := e.getData(url)
	if err != nil {
		return nil, err
	}

	stats := &NodesStats{}

	err = json.Unmarshal(body, stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (n NodesStats) String(dc func(format string, a ...interface{}) string) string {
	var buffer bytes.Buffer
	undr := color.New(color.Underline).SprintfFunc()
	for nodeID := range n.Nodes {
		buffer.WriteString("\n")
		buffer.WriteString(dc(undr(nodeID)))
		buffer.WriteString(dc(":"))
		buffer.WriteString(dc("%v", n.Nodes[nodeID]))
	}

	return buffer.String()
}

func (n NodeStats) String() string {
	return pad(fmt.Sprintf("\n"+
		"Name: %s\n"+
		"Transport Address: %s\n"+
		"Host: %s\n"+
		"IP: %v\n"+
		"JVM: %v\n"+
		"Indices: %v\n"+
		"Filesystem: %v",
		n.Name, n.TransportAddress, n.Host, strings.Join(n.IP, ", "), n.JVM, n.Indices, n.FS), 2)
}

func (i NodeIndicesStats) String() string {
	return pad(fmt.Sprintf(
		"\nDocuments:\n"+
			"  Count: %d\n"+
			"  Deleted: %d\n"+
			"Store:\n"+
			"  Size: %.1f Mb\n"+
			"  Throttle Time: %d sec",
		i.Docs.Count, i.Docs.Deleted, megabytes(i.Store.Size), i.Store.ThrottleTime/1000), 2)
}

func (j NodeJvmStats) String() string {
	return pad(fmt.Sprintf(
		"\nUptime: %s"+
			"\nMemory: %v"+
			"\nThreads: %v",
		fmtTime(j.Uptime/1000), j.Memory, j.Threads), 2)
}

func (t NodeThreadStats) String() string {
	return pad(fmt.Sprintf(
		"\nCount: %d"+
			"\nPeak Count: %d",
		t.Count, t.PeakCount), 2)
}

func (m NodeMemoryStats) String() string {
	return pad(fmt.Sprintf(
		"\n"+
			"Heap Used: %.1f Mb\n"+
			"Heap Committed: %.1f Mb\n"+
			"Heap Max: %.1f Mb\n"+
			"Non-Heap Used: %.1f Mb\n"+
			"Non-Heap Committed: %.1f Mb",
		megabytes(m.HeapUsed), megabytes(m.HeapCommitted), megabytes(m.HeapMax),
		megabytes(m.NonHeapUsed), megabytes(m.NonHeapCommitted)), 2)
}

func (f FsStats) String() string {
	var buffer bytes.Buffer
	for i, n := range f.Data {
		buffer.WriteString("\n    " + strconv.FormatInt(int64(i), 10) + ":")
		buffer.WriteString(pad(n.String(), 4))
		buffer.WriteString("\n")
	}
	data := buffer.String()
	return pad(fmt.Sprintf(
		"\nTotal:\n"+
			"  Total: %6.1f Mb\n"+
			"  Free: %6.1f Mb\n"+
			"  Available: %6.1f Mb\n"+
			"  Data: %v",
		megabytes(f.Total.Total), megabytes(f.Total.Free), megabytes(f.Total.Available),
		data), 2)
}

func (f FsNodeStats) String() string {
	return pad(fmt.Sprintf(
		"\n"+
			"Path: %s\n"+
			"Mount: %s\n"+
			"FS Type: %s\n"+
			"Total: %.1f Mb\n"+
			"Free: %.1f Mb\n"+
			"Available: %.1f Mb",
		f.Path, f.Mount, f.Type, megabytes(f.Total), megabytes(f.Free), megabytes(f.Available)), 2)
}

func megabytes(b int64) float64 {
	return float64(b) / 1024.0 / 1024.0
}

func fmtTime(seconds int64) string {
	minutes := seconds / 60
	seconds = seconds % 60
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%d:%d:%d", hours, minutes, seconds)
}

func pad(s string, num int) string {
	return strings.Replace(s, "\n", "\n"+strings.Repeat(" ", num), -1)
}
