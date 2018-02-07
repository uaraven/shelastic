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

type PingResponse struct {
	Status      int
	ClusterName string
	Version     string
}

type ClusterHealth struct {
	ClusterName string
	Status      string
}

type ShortNodeInfo struct {
	Name string
	Host string
	IP   string
}

type Es struct {
	esURL       *url.URL
	client      *http.Client
	ClusterName string
	version     []int
}

func Connect(host string) (*Es, *PingResponse, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Port() == "" {
		u.Host = u.Host + ":9200"
	}

	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	es := Es{esURL: u, client: client}

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
	}

	return &es, ping, err
}

func (e Es) get(path string) (*http.Response, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	resp, err := e.client.Get(reqURL.String())
	return resp, err
}

func (e Es) getJson(path string) (map[string]interface{}, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	resp, err := e.client.Get(reqURL.String())

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var body map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, err
	}
	return body, err
}

func (e Es) Ping() (*PingResponse, error) {
	body, err := e.getJson("/")

	if err != nil {
		return nil, err
	}

	return &PingResponse{
		Status:      int(body["status"].(float64)),
		ClusterName: body["cluster_name"].(string),
		Version:     body["version"].(map[string]interface{})["number"].(string),
	}, nil
}

func (e Es) Health() (*ClusterHealth, error) {
	body, err := e.getJson("/_cluster/health")

	if err != nil {
		return nil, err
	}

	return &ClusterHealth{
		ClusterName: body["cluster_name"].(string),
		Status:      body["status"].(string),
	}, nil
}

func (e Es) ListIndices() ([]string, error) {
	body, err := e.getJson("/_all")

	if err != nil {
		return nil, err
	}
	result := make([]string, len(body))
	idx := 0
	for index := range body {
		result[idx] = index
		idx++
	}
	return result, nil
}

func (e Es) ListNodes() ([]*ShortNodeInfo, error) {
	body, err := e.getJson("/_nodes")

	if err != nil {
		return nil, err
	}

	nodes := body["nodes"].(map[string]interface{})

	result := make([]*ShortNodeInfo, len(nodes))
	idx := 0
	for node := range nodes {
		nodeInfo := nodes[node].(map[string]interface{})

		sni := &ShortNodeInfo{
			Name: nodeInfo["name"].(string),
			Host: nodeInfo["host"].(string),
			IP:   nodeInfo["ip"].(string),
		}
		result[idx] = sni
		idx++
	}
	return result, nil
}

func (sni ShortNodeInfo) String() string {
	return fmt.Sprintf("%s @ %s[%s]", sni.Name, sni.Host, sni.IP)
}
