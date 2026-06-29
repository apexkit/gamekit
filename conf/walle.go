package conf

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/apexkit/gamekit/infra/walle"
)

const (
	envWalleURL   = "WALLE_URL"
	envWalleToken = "WALLE_TOKEN"
)

// ApplyByWalle overlays Bootstrap with resources from Walle when GROUP is set.
// WALLE_URL and WALLE_TOKEN must also be present. When GROUP is empty, this is a no-op.
func ApplyByWalle(bc *Bootstrap) error {
	if bc == nil {
		return fmt.Errorf("bootstrap is nil")
	}
	groupRaw := strings.TrimSpace(os.Getenv(walle.EnvGroup))
	if groupRaw == "" {
		return nil
	}

	walleURL := strings.TrimSpace(os.Getenv(envWalleURL))
	walleToken := strings.TrimSpace(os.Getenv(envWalleToken))
	if walleURL == "" || walleToken == "" {
		return fmt.Errorf("%s is set but %s and %s are required", walle.EnvGroup, envWalleURL, envWalleToken)
	}

	client := walle.NewClient(walleURL, walleToken)
	group, err := client.GetGameGroup(context.Background(), groupRaw)
	if err != nil {
		return fmt.Errorf("walle group=%q url=%s: %w", groupRaw, walleURL, err)
	}

	if group.MySQLConfig != nil {
		dsn, err := walle.MySQLDSN(group.MySQLConfig)
		if err != nil {
			return fmt.Errorf("mysql: %w", err)
		}
		bc.Data = ensureData(bc.Data)
		bc.Data.Database = []*Data_Database{{
			Name: "default",
			Type: "mysql",
			Dsn:  dsn,
		}}
	}

	if group.RedisConfig != nil {
		addr, password, db, useTLS, err := walle.RedisEndpoint(group.RedisConfig)
		if err != nil {
			return fmt.Errorf("redis: %w", err)
		}
		host, port, err := walle.SplitRedisAddr(addr)
		if err != nil {
			return err
		}
		bc.Data = ensureData(bc.Data)
		bc.Data.Redis = &Data_Redis{
			Host:     host,
			Port:     port,
			Password: password,
			Database: int32(db),
			UseTls:   useTLS,
		}
	}

	if group.ConsulConfig != nil {
		address, scheme, err := walle.ConsulEndpoint(group.ConsulConfig)
		if err != nil {
			return fmt.Errorf("consul: %w", err)
		}
		bc.Registry = ensureRegistry(bc.Registry)
		bc.Registry.Consul = &Registry_Consul{
			Address: address,
			Scheme:  scheme,
		}
	}

	if group.S3Config != nil {
		bc.Aws = &Aws{
			AccessKey: group.S3Config.AccessKey,
			SecretKey: group.S3Config.SecretKey,
			Region:    group.S3Config.Region,
			Bucket:    group.S3Config.Bucket,
		}
	}

	if group.NatsConfig != nil {
		url, err := walle.NatsEndpoint(group.NatsConfig)
		if err != nil {
			return fmt.Errorf("nats: %w", err)
		}
		bc.Data = ensureData(bc.Data)
		bc.Data.Eventbus = &Data_Eventbus{
			Type: "nats",
			Url:  url,
		}
	}

	if group.LogLevel != "" {
		bc.Log = ensureLog(bc.Log)
		bc.Log.Level = int32(walle.LogLevelZap(group.LogLevel))
	}

	return nil
}

func ensureData(data *Data) *Data {
	if data != nil {
		return data
	}
	return &Data{}
}

func ensureRegistry(registry *Registry) *Registry {
	if registry != nil {
		return registry
	}
	return &Registry{}
}

func ensureLog(logCfg *Log) *Log {
	if logCfg != nil {
		return logCfg
	}
	return &Log{}
}
