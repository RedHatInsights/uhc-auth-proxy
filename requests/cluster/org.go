package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var orgURL = "https://api.openshift.com/api/accounts_mgmt/v1/organizations/%s"

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

func GetOrg(r OrgRequest) (*Org, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(orgURL, r.OrgID), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.Token))
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%s\n", b)

	res := &Org{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, err
	}

	return res, nil
}
