package grpc

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type ConnectionPool struct {
	mu          sync.RWMutex
	connections map[string]*grpc.ClientConn
	registry    registry.Discovery
	serviceName string
	logger      *log.Helper
}

func NewConnectionPool(serviceName string, ds registry.Discovery, logger log.Logger) *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*grpc.ClientConn),
		registry:    ds,
		serviceName: serviceName,
		logger:      log.NewHelper(logger),
	}
}

func (cp *ConnectionPool) createOrGetConnection(name, version string) (*grpc.ClientConn, error) {
	key := name + "@" + version

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// 检查现有连接是否健康
	if conn, exists := cp.connections[key]; exists && cp.isConnectionHealthy(conn) {
		return conn, nil
	}

	// 清理不健康的连接
	if conn, exists := cp.connections[key]; exists {
		conn.Close()
		delete(cp.connections, key)
		cp.logger.Debugf("Removed unhealthy connection: %s", key)
	}

	// 多次尝试创建连接，提高连接到新服务的概率
	var conn *grpc.ClientConn
	var err error

	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 500 * time.Millisecond) // 0ms, 500ms, 1000ms
		}

		conn, err = GetGrpcConn(cp.serviceName, cp.registry, name, version)
		if err == nil {
			// 测试连接是否可用
			state := conn.GetState()
			if state == connectivity.Ready || state == connectivity.Idle {
				break
			}
			conn.Close()
			conn = nil
		}
	}

	if err != nil || conn == nil {
		cp.logger.Errorf("Failed to create gRPC connection to %s@%s after 3 attempts: %v", name, version, err)
		return nil, err
	}

	cp.connections[key] = conn
	cp.logger.Infof("Created gRPC connection: %s@%s", name, version)

	// 启动连接状态监控
	go cp.monitorConnection(key, conn)

	return conn, nil
}

func (cp *ConnectionPool) GetConnection(rpcConfig string) (*grpc.ClientConn, error) {
	// 解析RPC配置字符串，格式: "service_name@version"
	parts := strings.Split(rpcConfig, "@")
	if len(parts) < 2 {
		return nil, errors.New("invalid rpc format, expected: service_name@version")
	}

	name := parts[0]
	version := parts[1]

	return cp.createOrGetConnection(name, version)
}

func (cp *ConnectionPool) isConnectionHealthy(conn *grpc.ClientConn) bool {
	if conn == nil {
		return false
	}

	state := conn.GetState()
	// 只有Ready和Idle状态才认为是健康的
	// TransientFailure, Connecting, Shutdown 都认为不健康
	return state == connectivity.Ready || state == connectivity.Idle
}

func (cp *ConnectionPool) CloseConnection(name, version string) {
	key := name + "@" + version

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if conn, exists := cp.connections[key]; exists {
		conn.Close()
		delete(cp.connections, key)
		cp.logger.Debugf("Closed gRPC connection to %s@%s", name, version)
	}
}

func (cp *ConnectionPool) CloseAll() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for key, conn := range cp.connections {
		conn.Close()
		cp.logger.Debugf("Closed gRPC connection: %s", key)
	}

	cp.connections = make(map[string]*grpc.ClientConn)
}

func (cp *ConnectionPool) HealthCheck() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for key, conn := range cp.connections {
		if !cp.isConnectionHealthy(conn) {
			cp.logger.Debugf("Removing unhealthy connection: %s", key)
			conn.Close()
			delete(cp.connections, key)
		}
	}
}

func (cp *ConnectionPool) monitorConnection(key string, conn *grpc.ClientConn) {
	for {
		state := conn.GetState()

		// 如果连接关闭或失败，从连接池中移除
		if state == connectivity.Shutdown || state == connectivity.TransientFailure {
			cp.logger.Warnf("Connection %s failed, removing from pool", key)

			cp.mu.Lock()
			if storedConn, exists := cp.connections[key]; exists && storedConn == conn {
				delete(cp.connections, key)
			}
			cp.mu.Unlock()
			break
		}

		// 等待状态变化
		if !conn.WaitForStateChange(context.Background(), state) {
			cp.logger.Debugf("Connection %s monitor timeout", key)
			break
		}
	}
}

func (cp *ConnectionPool) StartHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			cp.HealthCheck()
		}
	}()
}
