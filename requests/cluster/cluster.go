package cluster

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// GetAccountID requests a cluster registration with the given Request
func GetAccountID(wrapper ClientWrapper, r Registration) (*ClusterRegistrationResponse, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", "https://api.openshift.com/api/accounts_mgmt/v1/cluster_registrations", buf)

	b, err := wrapper.Do(req)
	if err != nil {
		return nil, err
	}
	res := &ClusterRegistrationResponse{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, err
	}
	return res, nil
}
