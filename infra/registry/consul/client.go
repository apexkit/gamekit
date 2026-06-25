package consul

import (
	"fmt"

	"github.com/apexkit/gamekit/conf"
	"github.com/apexkit/gamekit/infra/walle"
	kratosconsul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	consulAPI "github.com/hashicorp/consul/api"
)

type registrarOptions struct {
	skipRegister func() bool
	healthCheck  bool
	kratosOpts   []kratosconsul.Option
}

func defaultRegistrarOptions() registrarOptions {
	return registrarOptions{
		skipRegister: walle.SkipConsulRegister,
		healthCheck:  false,
	}
}

// RegistrarOption configures NewRegistrar.
type RegistrarOption func(*registrarOptions)

// WithSkipRegister overrides when Consul registration is skipped.
// Pass nil to always register; default skips when walle.SkipConsulRegister is true.
func WithSkipRegister(fn func() bool) RegistrarOption {
	return func(o *registrarOptions) {
		o.skipRegister = fn
	}
}

// WithHealthCheck enables Consul health checks on registration.
func WithHealthCheck(enabled bool) RegistrarOption {
	return func(o *registrarOptions) {
		o.healthCheck = enabled
	}
}

// WithRegistrarOptions passes through kratos consul registry options.
func WithRegistrarOptions(opts ...kratosconsul.Option) RegistrarOption {
	return func(o *registrarOptions) {
		o.kratosOpts = append(o.kratosOpts, opts...)
	}
}

func shouldSkipRegister(fn func() bool) bool {
	if fn == nil {
		return false
	}
	return fn()
}

// NewAPIClient builds a hashicorp/consul API client from registry config.
func NewAPIClient(reg *conf.Registry) (*consulAPI.Client, error) {
	if reg == nil || reg.Consul == nil {
		return nil, fmt.Errorf("consul registry config is nil")
	}
	c := consulAPI.DefaultConfig()
	c.Address = reg.Consul.Address
	c.Scheme = reg.Consul.Scheme
	if reg.Consul.Token != "" {
		c.Token = reg.Consul.Token
	}
	return consulAPI.NewClient(c)
}

// NewRegistrar creates a Consul registrar, or nil when registration should be skipped.
func NewRegistrar(reg *conf.Registry, opts ...RegistrarOption) registry.Registrar {
	cfg := defaultRegistrarOptions()
	for _, opt := range opts {
		opt(&cfg)
	}
	if shouldSkipRegister(cfg.skipRegister) {
		return nil
	}

	cli, err := NewAPIClient(reg)
	if err != nil {
		panic(err)
	}
	kopts := append([]kratosconsul.Option{kratosconsul.WithHealthCheck(cfg.healthCheck)}, cfg.kratosOpts...)
	return kratosconsul.New(cli, kopts...)
}

// NewDiscoveryFromRegistry creates a Consul discovery client from registry config.
func NewDiscoveryFromRegistry(reg *conf.Registry, opts ...kratosconsul.Option) registry.Discovery {
	cli, err := NewAPIClient(reg)
	if err != nil {
		panic(err)
	}
	return NewDiscovery(cli, opts...)
}
