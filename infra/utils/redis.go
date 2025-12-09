package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitRedis 初始化 Redis 客户端
//
// 参数说明：
//   - addr: Redis 服务地址，格式 "host:port"，例如 "127.0.0.1:6379" 或 AWS/阿里云集群地址。
//   - password: Redis 访问密码，如果没有可以传空字符串 ""。
//   - db: Redis 数据库编号（0-15），常用 0。
//   - enableTls: 是否启用 TLS 加密连接，true 表示开启（推荐生产环境使用）。
//   - skipVerifyTls: 是否跳过 TLS 证书验证：
//     true  = 跳过服务端证书校验（⚠️ 仅测试环境使用，不安全）。
//     false = 严格验证证书（生产环境推荐）。
//   - caCertPath: 自定义 CA 根证书路径（.pem 文件），
//     如果 skipVerifyTls=false 且需要校验证书时必填；
//     如果为空字符串 ""，则使用系统默认 CA。
//
// 返回值：
//   - *redis.Client: 已初始化的 Redis 客户端。
//   - error: 如果连接失败或证书加载错误，会返回错误。
func InitRedis(addr, password string, db int, enableTls, skipVerifyTls bool, caCertPath string) (*redis.Client, error) {
	opt := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	if enableTls {
		tlsConfig := &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: skipVerifyTls,
		}

		// 如果没有跳过校验 && 提供了 caCertPath，就加载 CA 证书
		if !skipVerifyTls && caCertPath != "" {
			caCert, err := os.ReadFile(caCertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA cert file: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to append CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}

		opt.TLSConfig = tlsConfig
	}

	client := redis.NewClient(opt)

	// 测试连接，超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return client, nil
}

func InitRedisByDNS(dsn string) (*redis.Client, error) {
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	// 测试连接，超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return client, nil
}
