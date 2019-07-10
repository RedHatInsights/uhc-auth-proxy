package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
}

// HTTPWrapper manages the headers and auth required to speak
// with the auth service.  It also provides a convenience method
// to get the bytes from a request.
type HTTPWrapper struct {
	OfflineAccessToken string
}

// Wrapper provides a convenience method for getting bytes from
// a http request
type Wrapper interface {
	Do(req *http.Request) ([]byte, error)
}

// AddHeaders sets the client headers, including the auth token
func (c *HTTPWrapper) AddHeaders(req *http.Request, token string) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
}

// Do is a convenience wrapper that returns the response bytes
func (c *HTTPWrapper) Do(req *http.Request) ([]byte, error) {
	token, err := GetToken(c.OfflineAccessToken)
	if err != nil {
		return nil, err
	}
	c.AddHeaders(req, token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request to %s failed: %d %s", req.RequestURI, resp.StatusCode, resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
