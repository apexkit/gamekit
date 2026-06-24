package listen

import (
	"fmt"
	"log"
)

const (
	ProdHTTPAddr  = "0.0.0.0:8080"
	ProdGRPCAddr  = "0.0.0.0:9090"
	ProdPprofAddr = "0.0.0.0:6060"

	DefaultLocalHTTPPort  = 5002
	DefaultLocalGRPCPort  = 9000
	DefaultLocalPprofPort = 6060
)

// Addresses holds HTTP and gRPC listen addresses.
type Addresses struct {
	HTTP string
	GRPC string
}

type config struct {
	serviceName     string
	localHTTPPort   int
	localGRPCPort   int
	localPprofPort  int
	prodHTTPAddr    string
	prodGRPCAddr    string
	prodPprofAddr   string
}

// Option configures Normalize and PprofListenAddr.
type Option func(*config)

func WithServiceName(name string) Option {
	return func(c *config) {
		c.serviceName = name
	}
}

func defaultConfig() config {
	return config{
		localHTTPPort:  DefaultLocalHTTPPort,
		localGRPCPort:  DefaultLocalGRPCPort,
		localPprofPort: DefaultLocalPprofPort,
		prodHTTPAddr:   ProdHTTPAddr,
		prodGRPCAddr:   ProdGRPCAddr,
		prodPprofAddr:  ProdPprofAddr,
	}
}

// Normalize applies listen addresses for local and deployed environments.
// Local: resolve HTTP/GRPC from config with automatic port bump on conflict.
// Non-local: fixed production addresses.
func Normalize(addrs *Addresses, opts ...Option) {
	if addrs == nil {
		return
	}
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if IsLocal() {
		addrs.HTTP = resolveLocalAddr("http", addrs.HTTP, cfg.localHTTPPort, cfg.serviceName)
		addrs.GRPC = resolveLocalAddr("grpc", addrs.GRPC, cfg.localGRPCPort, cfg.serviceName)
		return
	}
	addrs.HTTP = cfg.prodHTTPAddr
	addrs.GRPC = cfg.prodGRPCAddr
}

// PprofListenAddr returns the pprof listen address.
// Local: default 6060 with port bump on conflict. Non-local: fixed 0.0.0.0:6060.
func PprofListenAddr(opts ...Option) string {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if IsLocal() {
		return resolveLocalAddr("pprof", "", cfg.localPprofPort, cfg.serviceName)
	}
	return cfg.prodPprofAddr
}

func resolveLocalAddr(name, preferred string, defaultPort int, serviceName string) string {
	addr, startPort, err := ResolveLocalListenAddr(preferred, defaultPort)
	if err != nil {
		panic(fmt.Sprintf("resolve local %s port from %d: %v", name, startPort, err))
	}
	if _, resolvedPort := ParseHostPort(addr, DefaultHost, defaultPort); resolvedPort != startPort {
		prefix := serviceName
		if prefix == "" {
			prefix = "service"
		}
		log.Printf("[%s] local %s port %d in use, using %d instead", prefix, name, startPort, resolvedPort)
	}
	return addr
}
