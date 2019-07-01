package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var orgURL = "https://api.openshift.com/api/accounts_mgmt/v1/organizations/%s"

// GetOrg retrieves organization details
func GetOrg(wrapper ClientWrapper, orgID string) (*Org, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(orgURL, orgID), nil)

	b, err := wrapper.Do(req)
	if err != nil {
		return nil, err
	}

	res := &Org{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, err
	}

	return res, nil
}
