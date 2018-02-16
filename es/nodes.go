package es

import (
	"encoding/json"
	"fmt"
	"strings"
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

type OSInfo struct {
	Name    string `json:"name"`
	Arch    string `json:"arch"`
	Version string `json:"version"`
	CPUs    int    `json:"allocated_processors"`
}

type JVMInfo struct {
	Version   string   `json:"version"`
	VMName    string   `json:"vm_name"`
	VMVersion string   `json:"vm_version"`
	VMVendor  string   `json:"vm_vendor"`
	Arguments []string `json:"input_arguments"`
}

type NodeEnvironmentInfo struct {
	OS  *OSInfo  `json:"os"`
	JVM *JVMInfo `json:"jvm"`
}

type NodesEnvironmentInfo struct {
	Nodes map[string]NodeEnvironmentInfo `json:"nodes"`
}

func (e Es) GetNodeEnvironmentInfo(nodeNames []string) (*NodesEnvironmentInfo, error) {
	nodestr := strings.Join(nodeNames, ",")
	var url string
	if len(nodeNames) > 0 {
		url = fmt.Sprintf("_nodes/%s", nodestr)
	} else {
		url = "_nodes/"
	}

	body, err := e.getData(url)
	if err != nil {
		return nil, err
	}

	result := &NodesEnvironmentInfo{}

	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
