package es

import (
	"context"
	"strings"

	"github.com/olivere/elastic"
)

type EsConnection interface {
	Health() *elastic.ClusterHealthResponse
	ListIndices() []string
}

type Es struct {
	host    string
	elastic *elastic.Client
}

var (
	ctx = context.Background()
)

func Connect(host string) (*Es, error) {
	if !strings.Contains(host, ":") {
		host = host + ":9200"
	}
	if !strings.Contains(host, "http") {
		host = "http://" + host
	}

	client, err := elastic.NewClient(elastic.SetURL(host))

	if err != nil {
		return nil, err
	}
	return &Es{host, client}, nil
}

func (e Es) Health() (*elastic.ClusterHealthResponse, error) {
	return e.elastic.ClusterHealth().Do(ctx)
}

func (e Es) ListIndices() ([]string, error) {
	return e.elastic.IndexNames()
}
func (e Es) ListNodes() ([]string, error) {
	nodes, err := e.elastic.NodesInfo().Do(ctx)
	if err != nil {
		return nil, err
	}
	var names []string
	for k := range nodes.Nodes {
		names = append(names, k)
	}
	return names, nil
}
