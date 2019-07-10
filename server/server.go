package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
	l "github.com/redhatinsights/uhc-auth-proxy/logger"
	"github.com/redhatinsights/uhc-auth-proxy/requests/client"
	"github.com/redhatinsights/uhc-auth-proxy/requests/cluster"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	l.InitLogger()
	log = l.Log.Named("server")
}

// returns the cluster id from the user agent string used by the support operator
// support-operator/commit cluster/cluster_id
func getClusterID(userAgent string) (string, error) {
	spl := strings.SplitN(userAgent, " ", 2)
	if !strings.HasPrefix(spl[0], `support-operator/`) {
		return "", fmt.Errorf("Invalid user-agent: %s", userAgent)
	}

	if !strings.HasPrefix(spl[1], `cluster/`) {
		return "", fmt.Errorf("Invalid user-agent: %s", userAgent)
	}

	return strings.TrimPrefix(spl[1], `cluster/`), nil
}

func getToken(authorizationHeader string) (string, error) {
	if !strings.HasPrefix(authorizationHeader, `Bearer `) {
		return "", fmt.Errorf("Not a bearer token: '%s'", authorizationHeader)
	}

	return strings.TrimPrefix(authorizationHeader, `Bearer `), nil
}

// RootHandler returns a handler that uses the given client and token
func RootHandler(wrapper client.Wrapper) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterID, err := getClusterID(r.Header.Get("user-agent"))
		if err != nil {
			log.Error("Failed to get the cluster id", zap.Error(err))
			w.WriteHeader(400)
			return
		}

		token, err := getToken(r.Header.Get("Authorization"))
		if err != nil {
			log.Error("Failed to get the token", zap.Error(err))
			w.WriteHeader(400)
			return
		}

		reg := cluster.Registration{
			ClusterID:          clusterID,
			AuthorizationToken: token,
		}

		ident, err := cluster.GetIdentity(wrapper, reg)
		if err != nil {
			log.Error("could not authenticate given the credentials", zap.Error(err))
			w.WriteHeader(401)
			return
		}

		b, err := json.Marshal(ident)
		if err != nil {
			log.Error("Failed to marshal identity", zap.Error(err))
			w.WriteHeader(400)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		log.Info(fmt.Sprintf("Responding with: %s", b))
		w.Write(b)
	}
}

// Start starts the server
func Start(offlineAccessToken string) {
	r := chi.NewRouter()
	r.Use(
		request_id.ConfiguredRequestID("x-rh-insights-request-id"),
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.StripSlashes,
	)

	log.Info(fmt.Sprintf("OAT = %s", offlineAccessToken))
	wrapper := &client.HTTPWrapper{
		OfflineAccessToken: offlineAccessToken,
	}

	handler := RootHandler(wrapper)

	r.Get("/", handler)
	r.Get("/api/uhc-auth-proxy/v1", handler)

	port := viper.GetInt64("SERVER_PORT")

	log.Info(fmt.Sprintf("Starting server on port %d", port))

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("server closed with error", zap.Error(err))
	}
}
