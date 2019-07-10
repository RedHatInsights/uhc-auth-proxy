package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// Registration is the document posted to cluster registration service
type Registration struct {
	ClusterID          string `json:"cluster_id"`
	AuthorizationToken string `json:"authorization_token"`
}

// Response is the format of the cluster registration response
type ClusterRegistrationResponse struct {
	ClusterID          string `json:"cluster_id"`
	AuthorizationToken string `json:"authorization_token"`
	AccountID          string `json:"account_id"`
	ExpiresAt          string `json:"expires_at"`
}

type Organization struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	HRef string `json:"href"`
	Name string `json:"name"`
}

type Account struct {
	ID           string       `json:"id"`
	Kind         string       `json:"kind"`
	HRef         string       `json:"href"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	Username     string       `json:"username"`
	Email        string       `json:"email"`
	Banned       bool         `json:"banned"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Organization Organization `json:"organization"`
}

type Org struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	HRef         string    `json:"href"`
	Name         string    `json:"name"`
	ExternalID   string    `json:"external_id"`
	EbsAccountID string    `json:"ebs_account_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type OrgRequest struct {
	Token string
	OrgID string
}

type Internal struct {
	OrgID string `json:"org_id"`
}

type Identity struct {
	AccountNumber string            `json:"account_number"`
	Type          string            `json:"type"`
	Internal      Internal          `json:"internal"`
	System        map[string]string `json:"system,omitempty"`
}

type FakeWrapper struct {
	GetAccountIDResponse *ClusterRegistrationResponse
	GetAccountResponse   *Account
	GetOrgResponse       *Org
}

func (f *FakeWrapper) Do(req *http.Request) ([]byte, error) {
	switch req.URL.String() {
	case viper.GetString("GET_ACCOUNTID_URL"):
		b, err := json.Marshal(f.GetAccountIDResponse)
		return b, err
	case fmt.Sprintf(viper.GetString("ACCOUNT_DETAILS_URL"), "123"):
		b, err := json.Marshal(f.GetAccountResponse)
		return b, err
	case fmt.Sprintf(viper.GetString("ORG_DETAILS_URL"), "123"):
		b, err := json.Marshal(f.GetOrgResponse)
		return b, err
	}
	return nil, fmt.Errorf("FakeClientWrapper failed to handle a case: %s", req.URL.String())
}
