package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redhatinsights/uhc-auth-proxy/requests/client"
)

// GetIdentity is a facade over all the steps required to get an Identity
func GetIdentity(wrapper client.Wrapper, r Registration) (*Identity, error) {
	rr, err := GetAccountID(wrapper, r)
	if err != nil {
		return nil, err
	}

	ar, err := GetAccount(wrapper, rr.AccountID)
	if err != nil {
		return nil, err
	}

	or, err := GetOrg(wrapper, ar.Organization.ID)
	if err != nil {
		return nil, err
	}

	return &Identity{
		AccountNumber: or.EbsAccountID,
		Type:          "system",
		Internal: Internal{
			OrgID: or.ExternalID,
		},
	}, nil
}

var GetAccountIDURL = "https://api.openshift.com/api/accounts_mgmt/v1/cluster_registrations"

// GetAccountID requests a cluster registration with the given Request
func GetAccountID(wrapper client.Wrapper, r Registration) (*ClusterRegistrationResponse, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", GetAccountIDURL, buf)

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

var AccountURL = "https://api.openshift.com/api/accounts_mgmt/v1/accounts/%s"

// GetAccount retrieves account details
func GetAccount(wrapper client.Wrapper, accountID string) (*Account, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(AccountURL, accountID), nil)

	b, err := wrapper.Do(req)
	if err != nil {
		return nil, err
	}

	res := &Account{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, err
	}
	return res, nil
}

var OrgURL = "https://api.openshift.com/api/accounts_mgmt/v1/organizations/%s"

// GetOrg retrieves organization details
func GetOrg(wrapper client.Wrapper, orgID string) (*Org, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(OrgURL, orgID), nil)

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
