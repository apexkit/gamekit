package resource

// Options declares infrastructure to install (Cola ResourceOpt style).
type Options struct {
	WithMysql bool
	WithRedis bool
	WithNats  bool
}
