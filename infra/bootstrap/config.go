package bootstrap

import (
	"errors"
	"os"
	"strings"

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
	path = strings.TrimSpace(path)
	if path == "" {
		path = "./configs"
	}

	// If the config path doesn't exist, initialize an empty bootstrap.
	// This enables container images or local runs that rely on Walle (GROUP)
	// to fully populate Data/Registry while keeping server/log defaults.
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &gkconf.Bootstrap{
				Server:   &gkconf.Server{},
				Data:     &gkconf.Data{},
				Log:      &gkconf.Log{},
				Registry: &gkconf.Registry{},
			}, nil
		}
		return nil, err
	}

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
