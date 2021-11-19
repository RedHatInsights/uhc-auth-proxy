package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
)

var (
	client = &http.Client{
		Timeout: viper.GetDuration("TIMEOUT_SECONDS") * time.Second,
	}
	requestTimes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "uhc_auth_proxy_request_time",
		Help: "Time spent waiting per request per url",
		Buckets: []float64{
			10,
			100,
			1000,
		},
	}, []string{"url"})
	acctMgmtRequest = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "uhc_auth_proxy_to_acct_mgmt_request_status",
		Help: "UHC auth proxy to account management request status",
	}, []string{"code"})
)

// HTTPWrapper manages the headers and auth required to speak
// with the auth service.  It also provides a convenience method
// to get the bytes from a request.
type HTTPWrapper struct{}

// Wrapper provides a convenience method for getting bytes from
// a http request
type Wrapper interface {
	Do(req *http.Request, label string, cluster_id string, authorization_token string) ([]byte, error)
}

// AddHeaders sets the client headers, including the auth token
func (c *HTTPWrapper) AddHeaders(req *http.Request, cluster_id string, authorization_token string) {
	req.Header.Add("Authorization", fmt.Sprintf("AccessToken %s:%s", cluster_id, authorization_token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
}

// Do is a convenience wrapper that returns the response bytes
func (c *HTTPWrapper) Do(req *http.Request, label string, cluster_id string, authorization_token string) ([]byte, error) {

	c.AddHeaders(req, cluster_id, authorization_token)
	start := time.Now()
	resp, err := client.Do(req)
	acctMgmtRequest.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()
	requestTimes.With(prometheus.Labels{"url": label}).Observe(time.Since(start).Seconds())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request to %s failed: %d %s", req.URL.String(), resp.StatusCode, resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
