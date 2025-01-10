package logger

import (
	"flag"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is an instance of the global zap.Logger
var Log *zap.Logger

// InitLogger initializes the logger
func InitLogger() *zap.Logger {
	if Log == nil {
		initLogConfig()
		logLevel := getLogLevel()

		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		loggerConfig := zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(logLevel),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    encoderConfig,
			InitialFields:    map[string]interface{}{"app": "uhc-auth-proxy"},
		}

		var options []zap.Option

		configureCloudwatch(&options, loggerConfig)

		logger, _ := loggerConfig.Build(options...)

		defer func() {
			if err := logger.Sync(); err != nil {
				fmt.Printf("Failed to sync logger: %v", err)
			}
		}()
		Log = logger
	}

	return Log
}

func getLogLevel() zapcore.Level {
	level := logConfig.GetString(LogLevel)
	var logLevel zapcore.Level
	err := logLevel.UnmarshalText([]byte(level))
	if err != nil {
		fmt.Printf("Error reading configured log level: %s. Defaulting to info.", err)
		logLevel = zapcore.InfoLevel
	}

	if flag.Lookup("test.v") != nil {
		logLevel = zapcore.FatalLevel
	}

	return logLevel
}

func configureCloudwatch(options *[]zap.Option, loggerConfig zap.Config) {
	if len(logConfig.GetString(CwAwsAccessKeyId)) > 0 {
		cloudwatch, err := getCloudwatchCore(loggerConfig)
		if err != nil {
			fmt.Printf("Error configuring cloudwatch integration: %s. Skipping cloudwatch integration.\n", err)
		} else {
			*options = append(*options, cloudwatch)
		}
	}
}
