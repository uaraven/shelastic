package es

import (
	"context"
	"strings"

	"github.com/olivere/elastic"
)

type EsConnection interface {
	Status() string
}

type Es struct {
	host    string
	elastic *elastic.Client
}

var (
	ctx = context.Background()
)

func Connect(host string) *Es {

	if !strings.Contains(host, ":") {
		host = host + ":9200"
	}
	if !strings.Contains(host, "http") {
		host = "http://" + host
	}

	client, err := elastic.NewClient(elastic.SetURL(host))

	if err != nil {
		// Handle error
		panic(err)
	}
	return &Es{host, client}
}

func (e Es) Health() *elastic.ClusterHealthResponse {
	response, error := e.elastic.ClusterHealth().Do(ctx)
	if error != nil {
		panic(error)
	}
	return response
}
