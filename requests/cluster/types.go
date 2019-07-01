package cluster

import "time"

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
