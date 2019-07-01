package access

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type response struct {
	AccessToken string `json:"access_token"`
}

var URL = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"

func Do(offlineAccessToken string) (string, error) {
	client := &http.Client{}
	resp, err := client.PostForm(URL, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"cloud-services"},
		"refresh_token": {offlineAccessToken},
	})
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	r := &response{}
	if err := json.Unmarshal(body, r); err != nil {
		return "", fmt.Errorf("failed to unmarshal: %s", body)
	}

	return r.AccessToken, nil
}
