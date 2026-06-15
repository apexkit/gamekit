package consul

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
)

// ServiceResolver builds service instances from Consul entries.
// Nomad-registered services only expose Address/Port (no grpc TaggedAddress),
// so we publish grpc:// endpoints for cross-host RPC discovery.
func ServiceResolver(ctx context.Context, entries []*api.ServiceEntry) []*registry.ServiceInstance {
	services := make([]*registry.ServiceInstance, 0, len(entries))
	for _, entry := range entries {
		var version string
		for _, tag := range entry.Service.Tags {
			ss := strings.SplitN(tag, "=", 2)
			if len(ss) == 2 && ss[0] == "version" {
				version = ss[1]
			}
		}

		endpoints := make([]string, 0)
		for scheme, addr := range entry.Service.TaggedAddresses {
			if scheme == "lan_ipv4" || scheme == "wan_ipv4" || scheme == "lan_ipv6" || scheme == "wan_ipv6" {
				continue
			}
			endpoints = append(endpoints, addr.Address)
		}
		if len(endpoints) == 0 && entry.Service.Address != "" && entry.Service.Port != 0 {
			endpoints = append(endpoints, fmt.Sprintf("grpc://%s:%d", entry.Service.Address, entry.Service.Port))
		}

		services = append(services, &registry.ServiceInstance{
			ID:        entry.Service.ID,
			Name:      entry.Service.Service,
			Metadata:  entry.Service.Meta,
			Version:   version,
			Endpoints: endpoints,
		})
	}
	return services
}
