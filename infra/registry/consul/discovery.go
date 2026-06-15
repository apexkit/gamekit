package consul

import (
	kratosconsul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	consulAPI "github.com/hashicorp/consul/api"
)

// NewDiscovery creates a Consul discovery client that understands Nomad-registered services.
func NewDiscovery(cli *consulAPI.Client, opts ...kratosconsul.Option) registry.Discovery {
	opts = append([]kratosconsul.Option{kratosconsul.WithServiceResolver(ServiceResolver)}, opts...)
	return kratosconsul.New(cli, opts...)
}
