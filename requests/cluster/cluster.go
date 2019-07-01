package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var client = &http.Client{}

// Registration is the document posted to cluster registration service
type Registration struct {
	ClusterID          string `json:"cluster_id"`
	AuthorizationToken string `json:"authorization_token"`
}

// Request is needed to call the cluster registration service
type AccountIDRequest struct {
	Token        string
	Registration Registration
}

// Response is the format of the cluster registration response
type Response struct {
	ClusterID          string `json:"cluster_id"`
	AuthorizationToken string `json:"authorization_token"`
	AccountID          string `json:"account_id"`
	ExpiresAt          string `json:"expires_at"`
}

// Do requests a cluster registration with the given Request
func GetAccountID(r AccountIDRequest) (*Response, error) {
	body, err := json.Marshal(r.Registration)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", "https://api.openshift.com/api/accounts_mgmt/v1/cluster_registrations", buf)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.Token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	res := &Response{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, err
	}
	return res, nil
}
