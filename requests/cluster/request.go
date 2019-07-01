package cluster

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var client = &http.Client{}

// HttpClientWrapper manages the headers and auth required to speak
// with the auth service.  It also provides a convenience method
// to get the bytes from a request.
type HTTPClientWrapper struct {
	Token string
}

// ClientWrapper provides a convenience method for getting bytes from
// a http request
type ClientWrapper interface {
	Do(req *http.Request) ([]byte, error)
}

// AddHeaders sets the client headers, including the auth token
func (c *HTTPClientWrapper) AddHeaders(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
}

// Do is a convenience wrapper that returns the response bytes
func (c *HTTPClientWrapper) Do(req *http.Request) ([]byte, error) {
	c.AddHeaders(req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return b, nil
}
