package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
