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

func (e Es) get(path string) (*http.Response, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	if e.Debug {
		fmt.Printf("Request: GET %s\n\n", reqURL.String())
	}
	resp, err := e.client.Get(reqURL.String())
	if e.Debug {
		if err != nil {
			fmt.Println("Response error: ", err.Error())
		} else {
			fmt.Println("Response:", resp)
		}
	}
	return resp, err
}

func (e Es) post(path string, data string) (*http.Response, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	var resp *http.Response
	req, err := http.NewRequest(http.MethodPost, reqURL.String(), strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	cl := strconv.FormatInt(int64(len(data)), 10)
	req.Header.Add("Content-Length", cl)

	if e.Debug {
		fmt.Println("Request: ", req)
	}

	resp, err = e.client.Do(req)

	if e.Debug {
		if err != nil {
			fmt.Println("Response error: ", err.Error())
		} else {
			fmt.Println("Response:", resp)
		}
	}

	return resp, err
}

func (e Es) requestWithBody(method string, path string, data string) (map[string]interface{}, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	var resp *http.Response
	req, err := http.NewRequest(method, reqURL.String(), strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	cl := strconv.FormatInt(int64(len(data)), 10)
	req.Header.Add("Content-Length", cl)
	req.Header.Add("Content-Type", "application/json")

	if e.Debug {
		dumpRequest(req, data)
	}

	resp, err = e.client.Do(req)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if e.Debug {
		if err != nil {
			fmt.Println("Response error: ", err.Error())
		} else {
			dumpResponse(resp, bodyBytes)
		}
	}

	if err != nil {
		return nil, err
	}

	var body map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, err
	}
	return body, err
}

func (e Es) putJSON(path string, data string) (map[string]interface{}, error) {
	return e.requestWithBody(http.MethodPut, path, data)
}

func (e Es) postJSON(path string, data string) (map[string]interface{}, error) {
	return e.requestWithBody(http.MethodPost, path, data)
}

func (e Es) getJSONWithBody(path string, data string) (map[string]interface{}, error) {
	return e.requestWithBody(http.MethodGet, path, data)
}

func (e Es) getData(path string) ([]byte, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	if e.Debug {
		fmt.Printf("Request: GET %s\n\n", reqURL.String())
	}
	resp, err := e.client.Get(reqURL.String())
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if e.Debug {
		if err != nil {
			fmt.Println("Response error: ", err.Error())
		} else {
			dumpResponse(resp, bodyBytes)
		}
	}

	return bodyBytes, nil
}

func (e Es) getJSON(path string) (map[string]interface{}, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	if e.Debug {
		fmt.Printf("Request: GET %s\n\n", reqURL.String())
	}
	resp, err := e.client.Get(reqURL.String())
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if e.Debug {
		if err != nil {
			fmt.Println("Response error: ", err.Error())
		} else {
			dumpResponse(resp, bodyBytes)
		}
	}

	var body map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, err
	}
	return body, err
}

func (e Es) delete(path string) (map[string]interface{}, error) {
	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	reqURL := e.esURL.ResolveReference(pathURL)
	if e.Debug {
		fmt.Printf("Request: GET %s\n\n", reqURL.String())
	}
	req, err := http.NewRequest(http.MethodDelete, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	if e.Debug {
		dumpRequest(req, "")
	}

	resp, err := e.client.Do(req)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if e.Debug {
		if err != nil {
			fmt.Println("Response error: ", err.Error())
		} else {
			dumpResponse(resp, bodyBytes)
		}
	}

	var body map[string]interface{}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, err
	}

	return body, err
}

func dumpRequest(req *http.Request, body string) {
	fmt.Println(req.Method, req.URL.String(), req.Proto)
	for key := range req.Header {
		fmt.Println(key, ":", req.Header.Get(key))
	}
	fmt.Println()
	fmt.Println(body)
	fmt.Println()
}

func dumpResponse(resp *http.Response, respBody []byte) {
	fmt.Println(resp.Status, resp.Proto)

	for key := range resp.Header {
		fmt.Println(key, ":", resp.Header.Get(key))
	}
	fmt.Println()

	fmt.Println(string(respBody))
	fmt.Println()
}
