package debug

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/go-kratos/kratos/v2/log"
)

type PprofServer struct {
	addr string
	log  *log.Helper
	srv  *http.Server
}

func NewPprofServer(addr string, logger log.Logger) *PprofServer {
	return &PprofServer{
		addr: addr,
		log:  log.NewHelper(logger),
	}
}

func (s *PprofServer) Start(ctx context.Context) error {
	_ = ctx
	if s.addr == "" {
		return nil
	}

	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: http.DefaultServeMux,
	}

	go func() {
		s.log.Infof("pprof server listening on %s", s.addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Errorf("pprof server listen failed: %v", err)
		}
	}()

	return nil
}

func (s *PprofServer) Stop(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}
