package logger

import (
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatch "github.com/RedHatInsights/cloudwatch-v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var getCloudwatchCore = func(loggerCfg zap.Config) (zap.Option, error) {
	key := logConfig.GetString(CwAwsAccessKeyId)
	secret := logConfig.GetString(CwAwsSecretAccessKey)
	region := logConfig.GetString(CwAwsRegion)
	group := logConfig.GetString(CwLogGroup)
	stream := logConfig.GetString(CwLogStream)

	client := cloudwatchlogs.New(cloudwatchlogs.Options{
		Region: region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(key, secret, "")),
	})

	cwGroup := cloudwatch.NewGroup(group, client)
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

func (w writerWrapper) Sync() error {
	return w.Flush()
}
