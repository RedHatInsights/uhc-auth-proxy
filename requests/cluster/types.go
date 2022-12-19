package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/redhatinsights/uhc-auth-proxy/requests/client"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// Registration is the document posted to cluster registration service
type Registration struct {
	ClusterID          string `json:"cluster_id"`
	AuthorizationToken string `json:"authorization_token"`
}

type Account struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	HRef         string    `json:"href"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Banned       bool      `json:"banned"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Organization Org       `json:"organization"`
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

// AccountError - holds error information returned from accounts endpoint if error occurred
type AccountError struct {
	Href        string `json:"href"`
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Code        string `json:"code"`
	OperationId string `json:"operation_id"`
	Reason      string `json:"reason"`
	Inner       error  `json:"-"`
}

func (a *AccountError) Error() string {
	return "[Reason: " + a.Reason + ", Code: " + a.Code + ", ID: " + a.ID + "]"
}

func (a *AccountError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("href", a.Href)
	enc.AddString("id", a.ID)
	enc.AddString("kind", a.Kind)
	enc.AddString("code", a.Code)
	enc.AddString("operation_id", a.OperationId)
	enc.AddString("reason", a.Reason)
	return nil
}

func (a *AccountError) Unwrap() error { return a.Inner }

func (a *AccountError) Verbose() []byte {
	v, _ := json.Marshal(a)
	return v
}

type Internal struct {
	OrgID string `json:"org_id"`
}

type Identity struct {
	AccountNumber string            `json:"account_number"`
	OrgID         string            `json:"org_id"`
	Type          string            `json:"type"`
	Internal      Internal          `json:"internal"`
	System        map[string]string `json:"system,omitempty"`
}

type FakeWrapper struct {
	GetAccountResponse *Account
}

func (f *FakeWrapper) Do(req *http.Request, label string, cluster_id string, authorization_token string) ([]byte, error) {
	switch req.URL.String() {
	case viper.GetString("CURRENT_ACCOUNT_URL"):
		b, err := json.Marshal(f.GetAccountResponse)
		return b, err
	}
	return nil, fmt.Errorf("FakeClientWrapper failed to handle a case: %s", req.URL.String())
}

type ErrorWrapper struct{}

func (e *ErrorWrapper) Do(req *http.Request, label string, cluster_id string, authorization_token string) ([]byte, error) {
	return nil, fmt.Errorf("errWrapper for: %s", req.URL.String())
}

type ErrorWithBodyWrapper struct {
	AccountError *AccountError
	StatusCode   int
}

func (e *ErrorWithBodyWrapper) Do(req *http.Request, label string, cluster_id string, authorization_token string) ([]byte, error) {
	bytes, _ := json.Marshal(e.AccountError)
	return bytes, &client.HttpError{
		Message:    "error message",
		StatusCode: e.StatusCode,
	}
}
