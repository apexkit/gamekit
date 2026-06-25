package bootstrap

import (
	gkconf "github.com/apexkit/gamekit/conf"
	"github.com/apexkit/gamekit/infra/metric"
	"github.com/go-kratos/kratos/v2/log"
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
	ProvideLogger,
	ProvideInstanceID,
	ProvideMeter,
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
