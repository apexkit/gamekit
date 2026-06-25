package bootstrap

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
)

// NewApp builds a kratos.App with standard metadata. Registrar is optional.
func NewApp(meta Meta, logger log.Logger, reg registry.Registrar, servers ...transport.Server) *kratos.App {
	opts := []kratos.Option{
		kratos.ID(meta.ID),
		kratos.Name(meta.Name),
		kratos.Version(meta.Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(servers...),
	}
	if reg != nil {
		opts = append(opts, kratos.Registrar(reg))
	}
	return kratos.New(opts...)
}
