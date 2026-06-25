package bootstrap

import (
	"context"
	"flag"
	"fmt"

	gkconf "github.com/apexkit/gamekit/conf"
	"github.com/apexkit/gamekit/infra/listen"
	"github.com/go-kratos/kratos/v2/log"
)

// Option configures service bootstrap.
type Option func(*Config)

// Config holds bootstrap inputs collected from options.
type Config struct {
	ServiceName string
	ServiceID   string
	Version     string
	ConfPath    string
	confFlag    *string
	Walle       bool
	Pprof       bool
	WithMysql   bool
	WithRedis   bool
}

// Runtime is the initialized bootstrap context passed into wire providers.
type Runtime struct {
	Bootstrap  *gkconf.Bootstrap
	Logger     log.Logger
	InstanceID string
}

// WithServiceName sets the kratos service name.
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithServiceID sets the service instance id (kratos.ID).
func WithServiceID(id string) Option {
	return func(c *Config) {
		c.ServiceID = id
	}
}

// WithVersion sets the service version.
func WithVersion(version string) Option {
	return func(c *Config) {
		c.Version = version
	}
}

// WithConfPath sets the config directory or file path (-conf).
func WithConfPath(path string) Option {
	return func(c *Config) {
		c.ConfPath = path
	}
}

// WithConfFlag binds bootstrap to a -conf flag pointer; read after flag.Parse.
func WithConfFlag(path *string) Option {
	return func(c *Config) {
		c.confFlag = path
	}
}

// WithWalle enables Walle overlay when GROUP is set (default true).
func WithWalle(enabled bool) Option {
	return func(c *Config) {
		c.Walle = enabled
	}
}

// WithPprof enables the debug pprof server (default true).
func WithPprof(enabled bool) Option {
	return func(c *Config) {
		c.Pprof = enabled
	}
}

// WithMysql declares mysql usage in main (wire pulls store providers when needed).
func WithMysql(enabled bool) Option {
	return func(c *Config) {
		c.WithMysql = enabled
	}
}

// WithRedis declares redis usage in main (wire pulls store providers when needed).
func WithRedis(enabled bool) Option {
	return func(c *Config) {
		c.WithRedis = enabled
	}
}

// NewConfig applies options onto defaults.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		ConfPath: "./configs",
		Walle:    true,
		Pprof:    true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	return cfg
}

// NewRuntime loads config, logger, and optional pprof from cfg.
func NewRuntime(cfg *Config) (*Runtime, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("bootstrap config is nil")
	}
	if cfg.ServiceName == "" {
		return nil, nil, fmt.Errorf("bootstrap: service name is required")
	}

	flag.Parse()
	confPath := cfg.ConfPath
	if cfg.confFlag != nil && *cfg.confFlag != "" {
		confPath = *cfg.confFlag
	}
	if confPath == "" {
		confPath = "./configs"
	}

	bc, err := LoadBootstrap(confPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.Walle {
		if err := PrepareBootstrap(bc, cfg.ServiceName); err != nil {
			return nil, nil, fmt.Errorf("prepare bootstrap: %w", err)
		}
	} else {
		gkconf.ApplyListenAddrs(bc.GetServer(), cfg.ServiceName)
	}

	meta := Meta{
		ID:      cfg.ServiceID,
		Name:    cfg.ServiceName,
		Version: cfg.Version,
	}
	lg, err := NewLogger(bc.GetLog().GetLevel(), listen.IsLocal(), meta)
	if err != nil {
		return nil, nil, fmt.Errorf("init logger: %w", err)
	}

	gkconf.SetConf(bc, lg.Kratos)

	var stopPprof func(context.Context) error
	if cfg.Pprof {
		stopPprof, err = StartPprof(context.Background(), cfg.ServiceName, lg.Kratos)
		if err != nil {
			lg.Sync()
			return nil, nil, fmt.Errorf("start pprof: %w", err)
		}
	}

	instanceID := fmt.Sprintf("%v(%v)-%v", cfg.ServiceName, cfg.Version, cfg.ServiceID)
	runtime := &Runtime{
		Bootstrap:  bc,
		Logger:     lg.Kratos,
		InstanceID: instanceID,
	}

	cleanup := func() {
		if stopPprof != nil {
			if err := stopPprof(context.Background()); err != nil {
				log.Errorf("shutdown pprof failed: %v", err)
			}
		}
		lg.Sync()
	}

	return runtime, cleanup, nil
}
