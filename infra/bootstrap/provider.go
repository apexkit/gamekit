package bootstrap

import (
	gkconf "github.com/apexkit/gamekit/conf"
	gkconsul "github.com/apexkit/gamekit/infra/registry/consul"
	"github.com/apexkit/gamekit/eventbus"
	"github.com/apexkit/gamekit/infra/metric"
	storemysql "github.com/apexkit/gamekit/infra/store/mysql"
	goredis "github.com/redis/go-redis/v9"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
)

// ProviderSet exposes bootstrap runtime pieces to wire.
var ProviderSet = wire.NewSet(
	NewConfig,
	NewRuntime,
	ProvideServer,
	ProvideData,
	ProvideRpc,
	ProvidePrometheus,
	ProvideRegistry,
	ProvideRegistrar,
	ProvideLogger,
	ProvideInstanceID,
	ProvideMeter,
	ProvideDatabases,
	ProvideRedisClient,
	ProvideEventBus,
)

func ProvideServer(rt *Runtime) *gkconf.Server {
	if rt == nil || rt.Bootstrap == nil {
		return nil
	}
	return rt.Bootstrap.Server
}

func ProvideData(rt *Runtime) *gkconf.Data {
	if rt == nil || rt.Bootstrap == nil {
		return nil
	}
	return rt.Bootstrap.Data
}

func ProvideRpc(rt *Runtime) *gkconf.Rpc {
	if rt == nil || rt.Bootstrap == nil {
		return nil
	}
	return rt.Bootstrap.Rpc
}

func ProvidePrometheus(rt *Runtime) *gkconf.Prometheus {
	if rt == nil || rt.Bootstrap == nil {
		return nil
	}
	return rt.Bootstrap.Prometheus
}

func ProvideRegistry(rt *Runtime) *gkconf.Registry {
	if rt == nil || rt.Bootstrap == nil {
		return nil
	}
	return rt.Bootstrap.Registry
}

// ProvideRegistrar creates a Consul registrar (nil when local+GROUP skips register).
func ProvideRegistrar(reg *gkconf.Registry) registry.Registrar {
	return gkconsul.NewRegistrar(reg)
}

func ProvideLogger(rt *Runtime) log.Logger {
	if rt == nil {
		return nil
	}
	return rt.Logger
}

func ProvideInstanceID(rt *Runtime) string {
	if rt == nil {
		return ""
	}
	return rt.InstanceID
}

// ProvideMeter creates the default Prometheus-backed Kratos server meter.
func ProvideMeter(instanceID string) *metric.KratosMeter {
	return metric.NewMeter(instanceID)
}

func ProvideDatabases(rt *Runtime) *storemysql.Databases {
	if rt == nil || rt.Resources == nil {
		return nil
	}
	return rt.Resources.Databases
}

func ProvideRedisClient(rt *Runtime) *goredis.Client {
	if rt == nil || rt.Resources == nil {
		return nil
	}
	return rt.Resources.Redis
}

func ProvideEventBus(rt *Runtime) eventbus.EventBus {
	if rt == nil || rt.Resources == nil {
		return nil
	}
	return rt.Resources.EventBus
}
