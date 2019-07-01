package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/redhatinsights/uhc-auth-proxy/pkg/requests/cluster"
)

// returns the cluster id from the user agent string used by the support operator
// support-operator/commit cluster/cluster_id
func getClusterID(userAgent string) (string, error) {
	spl := strings.SplitN(userAgent, " ", 2)
	if !strings.HasPrefix(spl[0], `support-operator/`) {
		return "", errors.New("Invalid user-agent")
	}

	if !strings.HasPrefix(spl[1], `cluster/`) {
		return "", errors.New("Invalid user-agent")
	}

	return strings.TrimPrefix(spl[1], `cluster/`), nil
}

func getToken(authorizationHeader string) (string, error) {
	if !strings.HasPrefix(authorizationHeader, `Bearer: `) {
		return "", errors.New("Not a bearer token")
	}

	return strings.TrimPrefix(authorizationHeader, `Bearer: `), nil
}

func main() {
	r := chi.NewRouter()
	r.Use(
		request_id.ConfiguredRequestID("x-rh-insights-request-id"),
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		clusterID, err := getClusterID(r.Header.Get("user-agent"))
		if err != nil {
			w.WriteHeader(400)
			return
		}

		token, err := getToken(r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(400)
			return
		}

		reg := &cluster.Registration{
			ClusterID:          clusterID,
			AuthorizationToken: token,
		}

		b, err := json.Marshal(reg)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		
	})
}
