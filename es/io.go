package es

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (e Es) get(path string) (*http.Response, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	resp, err := e.client.Get(reqURL.String())
	return resp, err
}

func (e Es) post(path string, data string) (*http.Response, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	var resp *http.Response
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	cl := strconv.FormatInt(int64(len(data)), 10)
	req.Header.Add("Content-Length", cl)

	resp, err = e.client.Do(req)

	return resp, err
}

func (e Es) getJSON(path string) (map[string]interface{}, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	resp, err := e.client.Get(reqURL.String())
	if err != nil {
		return nil, err
	}

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

func (e Es) getJSONInto(path string, container interface{}) error {
	pathURL, err := url.Parse(path)
	if err != nil {
		return err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	resp, err := e.client.Get(reqURL.String())
	if err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, &container)
}
