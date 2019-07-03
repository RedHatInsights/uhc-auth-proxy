package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var client = &http.Client{}

// HTTPWrapper manages the headers and auth required to speak
// with the auth service.  It also provides a convenience method
// to get the bytes from a request.
type HTTPWrapper struct {
	Token string
}

// Wrapper provides a convenience method for getting bytes from
// a http request
type Wrapper interface {
	Do(req *http.Request) ([]byte, error)
}

// AddHeaders sets the client headers, including the auth token
func (c *HTTPWrapper) AddHeaders(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
}

// Do is a convenience wrapper that returns the response bytes
func (c *HTTPWrapper) Do(req *http.Request) ([]byte, error) {
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

type FakeWrapper struct {
	GetAccountIDResponse *ClusterRegistrationResponse
	GetAccountResponse   *Account
	GetOrgResponse       *Org
}

func (f *FakeWrapper) Do(req *http.Request) ([]byte, error) {
	switch req.URL.String() {
	case GetAccountIDURL:
		b, err := json.Marshal(f.GetAccountIDResponse)
		return b, err
	case fmt.Sprintf(AccountURL, "123"):
		b, err := json.Marshal(f.GetAccountResponse)
		return b, err
	case fmt.Sprintf(OrgURL, "123"):
		b, err := json.Marshal(f.GetOrgResponse)
		return b, err
	}
	return nil, fmt.Errorf("FakeClientWrapper failed to handle a case: %s", req.URL.String())
}
