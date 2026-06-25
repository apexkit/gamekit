package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger creates a JSON production logger writing to stderr.
// DisableStacktrace keeps only kratos caller (file:line), not full stack traces.
func NewZapLogger(level int32) (*zap.Logger, error) {
	return newProdZapLogger(level)
}

// NewZapLoggerForEnv creates a dev console logger when local is true; otherwise production JSON.
func NewZapLoggerForEnv(level int32, local bool) (*zap.Logger, error) {
	if local {
		return newDevZapLogger(level)
	}
	return newProdZapLogger(level)
}

func newProdZapLogger(level int32) (*zap.Logger, error) {
	zcfg := zap.NewProductionConfig()
	zcfg.Level = zap.NewAtomicLevelAt(zapcore.Level(level))
	zcfg.Encoding = "json"
	zcfg.OutputPaths = []string{"stderr"}
	zcfg.ErrorOutputPaths = []string{"stderr"}
	zcfg.EncoderConfig.TimeKey = "timestamp"
	zcfg.EncoderConfig.MessageKey = "message"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zcfg.DisableStacktrace = true
	return zcfg.Build()
}

func newDevZapLogger(level int32) (*zap.Logger, error) {
	zcfg := zap.NewDevelopmentConfig()
	zcfg.Level = zap.NewAtomicLevelAt(zapcore.Level(level))
	zcfg.OutputPaths = []string{"stdout"}
	zcfg.ErrorOutputPaths = []string{"stdout"}
	zcfg.DisableStacktrace = true
	return zcfg.Build()
}
