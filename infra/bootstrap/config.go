package bootstrap

import (
	gkconf "github.com/apexkit/gamekit/conf"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
)

// LoadConfig loads yaml config from path into dest.
func LoadConfig(path string, dest any) error {
	c := config.New(
		config.WithSource(
			file.NewSource(path),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return err
	}
	return c.Scan(dest)
}

// LoadBootstrap loads the shared gamekit Bootstrap from yaml.
func LoadBootstrap(path string) (*gkconf.Bootstrap, error) {
	var bc gkconf.Bootstrap
	if err := LoadConfig(path, &bc); err != nil {
		return nil, err
	}
	return &bc, nil
}

// PrepareBootstrap applies Walle overlay and normalizes listen addresses.
func PrepareBootstrap(bc *gkconf.Bootstrap, serviceName string) error {
	if err := gkconf.ApplyByWalle(bc); err != nil {
		return err
	}
	gkconf.ApplyListenAddrs(bc.GetServer(), serviceName)
	return nil
}
