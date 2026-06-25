package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	gkconf "github.com/apexkit/gamekit/conf"
	"github.com/go-kratos/kratos/v2/log"
	goredis "github.com/redis/go-redis/v9"
)

// Options holds redis config from Bootstrap.data.
type Options struct {
	Data *gkconf.Data
}

// NewOptions builds redis store options from shared config.
func NewOptions(data *gkconf.Data) *Options {
	return &Options{Data: data}
}

// NewClient creates a go-redis client and verifies connectivity.
func NewClient(opt *Options, logger log.Logger) (*goredis.Client, func(), error) {
	if opt == nil || opt.Data == nil || opt.Data.Redis == nil {
		return nil, nil, fmt.Errorf("redis: data.redis is nil")
	}

	r := opt.Data.Redis
	addr := r.Host + ":" + strconv.Itoa(int(r.Port))
	clientOpt := &goredis.Options{
		Addr:     addr,
		Password: r.Password,
		DB:       int(r.Database),
	}
	useTLS := r.GetUseTls()
	log.NewHelper(logger).Infof("redis conf: addr=%s db=%d tls=%v", addr, clientOpt.DB, useTLS)
	if useTLS {
		clientOpt.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}
	}

	client := goredis.NewClient(clientOpt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.Ping(ctx).Result(); err != nil {
		_ = client.Close()
		endpoint := fmt.Sprintf("%s db=%d", addr, clientOpt.DB)
		hint := gkconf.DialHint("redis", endpoint, err)
		if hint != "" {
			return nil, nil, fmt.Errorf("redis ping failed (%s): %w; %s", endpoint, err, hint)
		}
		return nil, nil, fmt.Errorf("redis ping failed (%s): %w", endpoint, err)
	}

	log.NewHelper(logger).Infof("redis connected (endpoint=%s db=%d tls=%v)", addr, clientOpt.DB, clientOpt.TLSConfig != nil)
	cleanup := func() {
		_ = client.Close()
	}
	return client, cleanup, nil
}
