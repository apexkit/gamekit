package grpc

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"sync"
	"time"

	mmd "github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/filter"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel"
	google_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	selectorOnce sync.Once
)

func dialGrpcWithFilter(jobName, name string, ds registry.Discovery, f selector.NodeFilter) (*google_grpc.ClientConn, error) {
	const timeout = 10 * time.Second

	// 只设置一次全局selector，避免重复设置和并发问题
	selectorOnce.Do(func() {
		selector.SetGlobalSelector(wrr.NewBuilder())
	})

	meter := otel.Meter(jobName)
	metricRequests, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		return nil, err
	}
	metricSeconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		return nil, err
	}

	return grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+name),
		grpc.WithDiscovery(ds),
		grpc.WithTimeout(timeout),
		grpc.WithOptions(
			google_grpc.WithConnectParams(google_grpc.ConnectParams{
				MinConnectTimeout: 15 * time.Second,
			}),
			google_grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			}),
		),
		grpc.WithNodeFilter(f),
		grpc.WithMiddleware(
			mmd.Client(),
			tracing.Client(),
			metrics.Client(
				metrics.WithSeconds(metricSeconds),
				metrics.WithRequests(metricRequests),
			),
		),
	)
}

// 原函数：通过 version 过滤
func GetGrpcConn(jobName string, ds registry.Discovery, name string, version string) (*google_grpc.ClientConn, error) {
	return dialGrpcWithFilter(jobName, name, ds, filter.Version(version))
}

// 新函数：通过 Metadata["zone"] 过滤
func GetGrpcConnWithZone(jobName string, ds registry.Discovery, name string, zone string) (*google_grpc.ClientConn, error) {
	return dialGrpcWithFilter(jobName, name, ds, MetadataFilter("zone", zone))
}

// func GetGrpcConn(jobName string, ds registry.Discovery, name string, version string) (*google_grpc.ClientConn, error) {
// 	// 创建路由 Filter：筛选版本号为"2.0.0"的实例
// 	filter := filter.Version(version)
// 	// 由于 gRPC 框架的限制，只能使用全局 balancer name 的方式来注入 selector
// 	selector.SetGlobalSelector(wrr.NewBuilder())
// 	const timeout = 10 * time.Second

// 	meter := otel.Meter(jobName)
// 	var err error
// 	metricRequests, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	metricSeconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return grpc.DialInsecure(
// 		context.Background(),
// 		grpc.WithEndpoint("discovery:///"+name),
// 		grpc.WithDiscovery(ds),
// 		grpc.WithTimeout(timeout),
// 		// 启用健康检查和连接池配置
// 		grpc.WithOptions(
// 			google_grpc.WithConnectParams(google_grpc.ConnectParams{
// 				MinConnectTimeout: 15 * time.Second,
// 			}),
// 			google_grpc.WithKeepaliveParams(keepalive.ClientParameters{
// 				Time:                30 * time.Second,
// 				Timeout:             10 * time.Second,
// 				PermitWithoutStream: true,
// 			}),
// 		),
// 		// 通过 grpc.WithFilter 注入路由 Filter
// 		grpc.WithNodeFilter(filter),
// 		grpc.WithMiddleware(
// 			mmd.Client(),
// 			metrics.Client(
// 				metrics.WithSeconds(metricSeconds),
// 				metrics.WithRequests(metricRequests),
// 			),
// 		),
// 	)
// }
