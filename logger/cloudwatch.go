package logger

import (
	"github.com/RedHatInsights/cloudwatch"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
)

func getCloudwatchCore(loggerCfg zap.Config) (zap.Option, error) {
	key := logConfig.GetString(CwAwsAccessKeyId)
	secret := logConfig.GetString(CwAwsSecretAccessKey)
	region := logConfig.GetString(CwAwsRegion)
	group := logConfig.GetString(CwLogGroup)
	stream := logConfig.GetString(CwLogStream)

	awsConf := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(key, secret, "")).
		WithRegion(region)

	cloudWatchSession := session.Must(session.NewSession(awsConf))
	cloudWatchClient := cloudwatchlogs.New(cloudWatchSession)

	cwGroup := cloudwatch.NewGroup(group, cloudWatchClient)
	cwWriter, err := cwGroup.Create(stream)
	if err != nil {
		return nil, err
	}

	cwCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(loggerCfg.EncoderConfig),
		wrapWriter(cwWriter),
		loggerCfg.Level,
	)

	cwOption := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, cwCore)
	})

	return cwOption, nil
}

func wrapWriter(w io.Writer) zapcore.WriteSyncer {
	switch w := w.(type) {
	case *cloudwatch.Writer:
		return &writerWrapper{w}
	default:
		return zapcore.AddSync(w)
	}
}

type writerWrapper struct {
	*cloudwatch.Writer
}

// Sync - this method is required by zapcore
func (w writerWrapper) Sync() error {
	return w.Flush()
}
