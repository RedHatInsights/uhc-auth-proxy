package logger

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prashantv/gostub"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ = Describe("InitLogger", func() {
	AfterEach(func() {
		resetLogger()
	})

	It("Configures logger with cloudwatch", func() {
		// given
		stubs := gostub.New()
		stubs.SetEnv(CwAwsAccessKeyId, "access-key-id")
		stubs.SetEnv(CwAwsSecretAccessKey, "secret-access-key")

		var option = zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, zapcore.NewNopCore())
		})
		stubs.StubFunc(&getCloudwatchCore, option, nil)

		defer stubs.Reset()

		// when
		logger := InitLogger()

		// then
		Expect(logger.Core()).To(BeAssignableToTypeOf(getMultiCore()))
	})

	It("Configures logger without cloudwatch when access key not set", func() {
		// given
		stubs := gostub.New()
		stubs.SetEnv(CwAwsAccessKeyId, "")

		defer stubs.Reset()

		// when
		logger := InitLogger()

		// then
		Expect(logger.Core()).To(BeAssignableToTypeOf(getIoCore()))
	})
})

func resetLogger() {
	Log = nil
}

func getIoCore() zapcore.Core {
	l, _ := zap.NewDevelopment()
	return l.Core()
}

func getMultiCore() zapcore.Core {
	zap.NewNop()
	return zapcore.NewTee(zapcore.NewNopCore(), zapcore.NewNopCore())
}
