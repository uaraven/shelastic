package es

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PingResponse contains cluster name and ES version - response to ping command
type PingResponse struct {
	ClusterName string
	Version     string
}

// Es holds connection information for Elasticsearch cluster
type Es struct {
	host        string
	esURL       *url.URL
	client      *http.Client
	ClusterName string
	version     []int
	aliases     map[string]string
	nodes       map[string]*ShortNodeInfo
	Debug       bool
	ActiveIndex string
}

// Connect initiates connection to an Elasticsearch cluster node specified by host argument
func Connect(host string) (*Es, *PingResponse, error) {
	if !strings.Contains(host, ":") {
		host = host + ":9200"
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	u, err := url.Parse(host)
	if err != nil {
		return nil, nil, err
	}

	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	es := Es{client: client, esURL: u, host: host, ActiveIndex: ""}

	ping, err := es.Ping()

	if err == nil {
		vs := strings.Split(ping.Version, ".")
		var ver []int
		for _, v := range vs {
			vi, err := strconv.Atoi(v)
			if err != nil {
				return nil, nil, fmt.Errorf("Failed to parse Elasticsearch version %s", ping.Version)
			}
			ver = append(ver, vi)
		}
		es.version = ver
		es.ClusterName = ping.ClusterName
	} else {
		return nil, nil, err
	}

	aliases, err := es.buildAliasCache()

	if err != nil {
		return nil, nil, err
	}
	es.aliases = aliases

	nodes, err := es.ListNodes()
	if err != nil {
		return nil, nil, err
	}
	es.nodes = make(map[string]*ShortNodeInfo)
	for _, node := range nodes {
		es.nodes[node.UUID] = node
	}

	return &es, ping, err
}

// Ping performs ping request to an ES node
func (e Es) Ping() (*PingResponse, error) {
	body, err := e.getJSON("/")

	if err != nil {
		return nil, err
	}

	return &PingResponse{
		ClusterName: body["cluster_name"].(string),
		Version:     body["version"].(map[string]interface{})["number"].(string),
	}, nil
}

// Health returns current ClusterHealth
func (e Es) Health() (*ClusterHealth, error) {
	body, err := e.get("/_cluster/health")

	if err != nil {
		return nil, err
	}

	var data ClusterHealth
	bodyBytes, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
