package es

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// ClusterNode contains basic information about a cluster node
type ClusterNode struct {
	Name             string
	TransportAddress string
}

// ClusterState contains ES cluster state
type ClusterState struct {
	ClusterName string
	Version     int
	MasterNode  string
}

// ClusterHealth holds cluster health information
type ClusterHealth struct {
	ClusterName             string `json:"cluster_name" yaml:"Cluster Name"`
	Status                  string `json:"status" yaml:"Status"`
	NodeCount               int    `json:"number_of_nodes" yaml:"Nodes"`
	DataNodeCount           int    `json:"number_of_data-nodes" yaml:"Data Nodes"`
	ActiveShards            int    `json:"active_shards" yaml:"Active Shards"`
	ActivePrimaryShards     int    `json:"active_primary_shards" yaml:"Active Primary Shards"`
	RelocatingShards        int    `json:"relocating_shards" yaml:"Relocating Shards"`
	InitializingShards      int    `json:"initializing_shards" yaml:"Initializing Shards"`
	UnassignedShards        int    `json:"unassigned_shards" yaml:"Unassigned Shards"`
	DelayedUnassignedShards int    `json:"delayed_unassigned_shards" yaml:"Delayed Unassigned Shards"`
	PendingTasks            int    `json:"number_of_pending_tasks" yaml:"Pending Tasks"`
	InFlightFetch           int    `json:"number_of_in_flight_fetch" yaml:"In Flight Fetches"`
}

// ShortNodeInfo holds minimal node information
type ShortNodeInfo struct {
	UUID             string
	Name             string
	Host             string
	IP               string
	TransportAddress string
}

// ListNodes returns slice of *ShortNodeInfo structs containing node information
func (e Es) ListNodes() ([]*ShortNodeInfo, error) {
	body, err := e.getJSON("/_nodes")

	if err != nil {
		return nil, err
	}

	nodes := body["nodes"].(map[string]interface{})

	result := make([]*ShortNodeInfo, len(nodes))
	idx := 0
	for node := range nodes {
		nodeInfo := nodes[node].(map[string]interface{})

		sni := &ShortNodeInfo{
			UUID:             node,
			Name:             nodeInfo["name"].(string),
			TransportAddress: nodeInfo["transport_address"].(string),
			Host:             nodeInfo["host"].(string),
			IP:               nodeInfo["ip"].(string),
		}
		result[idx] = sni
		idx++
	}
	return result, nil
}

func (sni ShortNodeInfo) String() string {
	return fmt.Sprintf("%s @ %s [%s]", sni.Name, sni.Host, sni.IP)
}

func (h ClusterHealth) String() string {
	y, err := yaml.Marshal(h)
	if err != nil {
		return "Error"
	}
	return string(y)
}
