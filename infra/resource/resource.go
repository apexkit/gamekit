package resource

import (
	"github.com/apexkit/gamekit/eventbus"
	storemysql "github.com/apexkit/gamekit/infra/store/mysql"
	goredis "github.com/redis/go-redis/v9"
)

// Resource holds installed infrastructure clients (singleton per process).
type Resource struct {
	opts      *Options
	Databases *storemysql.Databases
	Redis     *goredis.Client
	EventBus  eventbus.EventBus
	cleanups  []func()
}

// Options returns install flags.
func (r *Resource) Options() *Options {
	if r == nil {
		return nil
	}
	return r.opts
}

// Close releases installed clients in reverse order.
func (r *Resource) Close() {
	if r == nil {
		return
	}
	for i := len(r.cleanups) - 1; i >= 0; i-- {
		if r.cleanups[i] != nil {
			r.cleanups[i]()
		}
	}
	r.cleanups = nil
}
