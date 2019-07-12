package client

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("ACCESS_TOKEN_URL", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token")
}