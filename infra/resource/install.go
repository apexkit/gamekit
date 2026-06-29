package resource

import (
	"fmt"
	"strings"

	gkconf "github.com/apexkit/gamekit/conf"
	"github.com/apexkit/gamekit/eventbus"
	storemysql "github.com/apexkit/gamekit/infra/store/mysql"
	storeredis "github.com/apexkit/gamekit/infra/store/redis"
	"github.com/go-kratos/kratos/v2/log"
)

// Install loads config, dials declared infrastructure, and returns a ready Resource.
// Fails fast when a declared resource cannot connect (Cola InstallResource semantics).
func Install(opts *Options, bc *gkconf.Bootstrap, logger log.Logger) (*Resource, error) {
	if opts == nil {
		return nil, fmt.Errorf("resource: options is nil")
	}
	if err := validate(opts, bc); err != nil {
		return nil, err
	}

	r := &Resource{opts: opts}
	data := bc.GetData()
	helper := log.NewHelper(logger)

	if opts.WithMysql {
		dbs, cleanup, err := storemysql.NewDatabases(storemysql.NewOptions(data), logger)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("resource mysql: %w", err)
		}
		r.Databases = dbs
		r.cleanups = append(r.cleanups, cleanup)
		helper.Info("resource: mysql installed")
	}

	if opts.WithRedis {
		client, cleanup, err := storeredis.NewClient(storeredis.NewOptions(data), logger)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("resource redis: %w", err)
		}
		r.Redis = client
		r.cleanups = append(r.cleanups, cleanup)
		helper.Info("resource: redis installed")
	}

	if opts.WithNats {
		eb := data.GetEventbus()
		bus, err := eventbus.NewConnectedNatsBus(eb.GetUrl())
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("resource nats: %w", err)
		}
		r.EventBus = bus
		r.cleanups = append(r.cleanups, func() { _ = bus.Close() })
		helper.Infof("resource: nats installed (%s)", eb.GetUrl())
	}

	return r, nil
}

func validate(opts *Options, bc *gkconf.Bootstrap) error {
	if bc == nil {
		return fmt.Errorf("resource: bootstrap is nil")
	}
	data := bc.GetData()

	if opts.WithMysql {
		if data == nil || len(data.GetDatabase()) == 0 {
			return fmt.Errorf("resource: WithMysql requires data.database")
		}
	}
	if opts.WithRedis {
		if data == nil || data.GetRedis() == nil {
			return fmt.Errorf("resource: WithRedis requires data.redis")
		}
	}
	if opts.WithNats {
		eb := data.GetEventbus()
		if eb == nil || eb.GetType() != "nats" || strings.TrimSpace(eb.GetUrl()) == "" {
			return fmt.Errorf("resource: WithNats requires data.eventbus type=nats with url")
		}
	}
	return nil
}
