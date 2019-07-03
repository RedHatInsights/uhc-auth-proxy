package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"sync"
	"time"
)

type response struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

var URL = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
var mutex = &sync.Mutex{}
var token = ""
var expires = time.Now().Unix()

func fetch(offlineAccessToken string) (*response, error) {
	resp, err := client.PostForm(URL, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {"cloud-services"},
		"refresh_token": {offlineAccessToken},
	})
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	r := &response{}
	if err := json.Unmarshal(body, r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %s", body)
	}

	return r, nil
}

// GetToken retrieves an access token from cache or the sso service
func GetToken(offlineAccessToken string) (string, error) {
	now := time.Now().Unix()
	if now >= expires || token == "" {
		r, err := fetch(offlineAccessToken)
		if err != nil {
			return "", err
		}

		mutex.Lock()
		token = r.AccessToken
		expires = now + r.ExpiresIn
		mutex.Unlock()
	}
	return token, nil
}
