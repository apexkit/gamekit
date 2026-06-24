package walle

// Response is the unified Walle OpenAPI envelope.
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    []GameGroup `json:"data"`
}

// GameGroup is a row from GET /openapi/game/group.
type GameGroup struct {
	GroupName            string        `json:"group_name"`
	LogLevel             string        `json:"log_level"`
	ServiceDiscoveryType string        `json:"service_discovery_type"`
	CreatedAt            string        `json:"created_at"`
	UpdatedAt            string        `json:"updated_at"`
	MySQLConfig          *MySQLConfig  `json:"mysql_config"`
	RedisConfig          *RedisConfig  `json:"redis_config"`
	ConsulConfig         *ConsulConfig `json:"consul_config"`
	S3Config             *S3Config     `json:"s3_config"`
}

type MySQLConfig struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	InternalHost   string `json:"internal_host"`
	ExternalHost   string `json:"external_host"`
	Port           int    `json:"port"`
	ConnectionType int    `json:"connection_type"`
	Account        string `json:"account"`
	Password       string `json:"password"`
	Database       string `json:"database"`
}

type RedisConfig struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	InternalEndpoint string `json:"internal_endpoint"`
	ExternalEndpoint string `json:"external_endpoint"`
	ConnectionType   int    `json:"connection_type"`
	AuthMode         string `json:"auth_mode"`
	RedisUser        string `json:"redis_user"`
	Auth             string `json:"auth"`
	DB               int    `json:"db"`
}

type ConsulConfig struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	InternalEndpoint string `json:"internal_endpoint"`
	ExternalEndpoint string `json:"external_endpoint"`
	ConnectionType   int    `json:"connection_type"`
}

type S3Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
}
