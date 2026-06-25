package bootstrap

import (
	gklog "github.com/apexkit/gamekit/infra/log"
	kratoszap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.uber.org/zap"
)

// Meta identifies a running service instance for logs and kratos.App.
type Meta struct {
	ID      string
	Name    string
	Version string
}

// Logger holds zap and kratos loggers created during bootstrap.
type Logger struct {
	Zap    *zap.Logger
	Kratos log.Logger
}

// NewLogger builds zap + kratos loggers with standard service fields.
func NewLogger(level int32, local bool, meta Meta) (*Logger, error) {
	z, err := gklog.NewZapLoggerForEnv(level, local)
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(z)

	base := kratoszap.NewLogger(z)
	log.SetLogger(base)

	kratosLogger := log.With(base,
		"caller", log.DefaultCaller,
		"service.id", meta.ID,
		"service.name", meta.Name,
		"service.version", meta.Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	return &Logger{Zap: z, Kratos: kratosLogger}, nil
}

// Sync flushes zap buffers; safe to call from defer.
func (l *Logger) Sync() {
	if l == nil || l.Zap == nil {
		return
	}
	_ = l.Zap.Sync()
}
