package logger

import (
	"github.com/spf13/viper"
	"os"
)

const (
	LogLevel             = "LOG_LEVEL"
	CwLogGroup           = "CW_LOG_GROUP"
	CwAwsRegion          = "CW_AWS_REGION"
	CwLogStream          = "CW_LOG_STREAM"
	CwAwsAccessKeyId     = "CW_AWS_ACCESS_KEY_ID"
	CwAwsSecretAccessKey = "CW_AWS_SECRET_ACCESS_KEY"
)

var logConfig *viper.Viper

func initLogConfig() {
	hostname, _ := os.Hostname()

	logConfig = viper.New()

	logConfig.SetDefault("LOG_LEVEL", "INFO")
	logConfig.SetDefault("CW_LOG_GROUP", "platform-dev")
	logConfig.SetDefault("CW_AWS_REGION", "us-east-1")
	logConfig.SetDefault("CW_LOG_STREAM", hostname)
	logConfig.AutomaticEnv()
}
