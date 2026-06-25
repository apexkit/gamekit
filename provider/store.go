package provider

import (
	"github.com/apexkit/gamekit/infra/store/mysql"
	"github.com/apexkit/gamekit/infra/store/redis"
	"github.com/google/wire"
)

// StoreProviderSet wires generic mysql/redis store constructors (atreus-style).
var StoreProviderSet = wire.NewSet(
	mysql.NewOptions,
	mysql.NewDatabases,
	redis.NewOptions,
	redis.NewClient,
)
