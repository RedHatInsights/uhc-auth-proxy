package client

import "github.com/spf13/viper"

func init() {
	viper.AutomaticEnv() //Ensure Env vars are included in viper.Get calls
	viper.SetDefault("ACCESS_TOKEN_URL", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token")
	viper.SetDefault("TIMEOUT_SECONDS", 30)
}
