package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	l "github.com/redhatinsights/uhc-auth-proxy/logger"
	"github.com/redhatinsights/uhc-auth-proxy/requests/client"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	l.InitLogger()
	log = l.Log.Named("cluster")
}

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
		Type:          "System",
		System: map[string]string{
			"cluster_id": r.ClusterID,
		},
		Internal: Internal{
			OrgID: or.ExternalID,
		},
	}, nil
}

// GetAccountID requests a cluster registration with the given Request
func GetAccountID(wrapper client.Wrapper, r Registration) (*ClusterRegistrationResponse, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	URL := viper.GetString("GET_ACCOUNTID_URL")
	req, err := http.NewRequest("POST", URL, buf)
	if err != nil {
		return nil, err
	}

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

// GetAccount retrieves account details
func GetAccount(wrapper client.Wrapper, accountID string) (*Account, error) {
	URL := viper.GetString("ACCOUNT_DETAILS_URL")
	req, _ := http.NewRequest("GET", fmt.Sprintf(URL, accountID), nil)

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

// GetOrg retrieves organization details
func GetOrg(wrapper client.Wrapper, orgID string) (*Org, error) {
	URL := viper.GetString("ORG_DETAILS_URL")
	req, _ := http.NewRequest("GET", fmt.Sprintf(URL, orgID), nil)

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
