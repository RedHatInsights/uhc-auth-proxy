package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
)

// HTTPWrapper manages the headers and auth required to speak
// with the auth service.  It also provides a convenience method
// to get the bytes from a request.
type HTTPWrapper struct {
	OfflineAccessToken string
}

// Wrapper provides a convenience method for getting bytes from
// a http request
type Wrapper interface {
	Do(req *http.Request) ([]byte, error)
}

// AddHeaders sets the client headers, including the auth token
func (c *HTTPWrapper) AddHeaders(req *http.Request, token string) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
}

// Do is a convenience wrapper that returns the response bytes
func (c *HTTPWrapper) Do(req *http.Request) ([]byte, error) {
	token, err := GetToken(c.OfflineAccessToken)
	if err != nil {
		return nil, err
	}
	c.AddHeaders(req, token)
	start := time.Now()
	resp, err := client.Do(req)
	requestTimes.With(prometheus.Labels{"url": req.URL.String()}).Observe(time.Since(start).Seconds())
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
