package bootstrap

import (
	"context"

	common_debug "github.com/apexkit/gamekit/infra/debug"
	"github.com/apexkit/gamekit/infra/listen"
	"github.com/go-kratos/kratos/v2/log"
)

// StartPprof starts the debug pprof server using listen.PprofListenAddr.
// The returned stop function should be deferred for graceful shutdown.
func StartPprof(ctx context.Context, serviceName string, logger log.Logger) (func(context.Context) error, error) {
	srv := common_debug.NewPprofServer(listen.PprofListenAddr(listen.WithServiceName(serviceName)), logger)
	if err := srv.Start(ctx); err != nil {
		return nil, err
	}
	return srv.Stop, nil
}
