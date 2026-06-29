package provider

import (
	"github.com/apexkit/gamekit/infra/store/mysql"
	"github.com/apexkit/gamekit/infra/store/redis"
	"github.com/google/wire"
)

// StoreProviderSet is deprecated: use gkbootstrap.ProviderSet with WithMysql/WithRedis in main.
// Resources are installed in bootstrap.NewRuntime via resource.Install.
var StoreProviderSet = wire.NewSet(
	mysql.NewOptions,
	mysql.NewDatabases,
	redis.NewOptions,
	redis.NewClient,
)
