package logger

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cww "github.com/lzap/cloudwatchwriter2"
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

	cwWriter, err := cww.NewWithClient(client, 500*time.Millisecond, group, stream)
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

func wrapWriter(w *cww.CloudWatchWriter) zapcore.WriteSyncer {
	return &writerWrapper{w}
}

type writerWrapper struct {
	*cww.CloudWatchWriter
}

func (w writerWrapper) Sync() error {
	w.Flush()
	return nil
}
