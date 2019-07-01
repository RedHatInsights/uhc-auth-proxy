package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

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

var accountURL = "https://api.openshift.com/api/accounts_mgmt/v1/accounts/%s"

type AccountRequest struct {
	Token     string
	AccountID string
}

func GetAccount(r AccountRequest) (*Account, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(accountURL, r.AccountID), nil)
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
	res := &Account{}
	if err := json.Unmarshal(b, res); err != nil {
		return nil, err
	}
	return res, nil
}
