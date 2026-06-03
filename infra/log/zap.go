package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger 创建 JSON 格式、输出到 stderr 的生产环境 logger。
// DisableStacktrace 仅保留 kratos caller（文件:行号），不输出 stacktrace。
func NewZapLogger(level int32) (*zap.Logger, error) {
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
