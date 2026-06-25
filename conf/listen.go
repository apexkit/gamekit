package conf

import (
	"github.com/apexkit/gamekit/infra/listen"
)

// ApplyListenAddrs normalizes HTTP/gRPC listen addresses via gamekit listen.
func ApplyListenAddrs(s *Server, serviceName string) {
	if s == nil {
		return
	}
	addrs := listen.Addresses{}
	if s.Http != nil {
		addrs.HTTP = s.Http.Addr
	}
	if s.Grpc != nil {
		addrs.GRPC = s.Grpc.Addr
	}
	listen.Normalize(&addrs, listen.WithServiceName(serviceName))
	if s.Http == nil {
		s.Http = &Server_HTTP{}
	}
	s.Http.Addr = addrs.HTTP
	if s.Grpc == nil {
		s.Grpc = &Server_GRPC{}
	}
	s.Grpc.Addr = addrs.GRPC
}
