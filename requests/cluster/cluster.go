package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetIdentity is a facade over all the steps required to get an Identity
func GetIdentity(wrapper ClientWrapper, r Registration) (*Identity, error) {
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

var accountURL = "https://api.openshift.com/api/accounts_mgmt/v1/accounts/%s"

// GetAccount retrieves account details
func GetAccount(wrapper ClientWrapper, accountID string) (*Account, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(accountURL, accountID), nil)

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
