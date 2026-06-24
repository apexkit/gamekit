package listen

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const DefaultHost = "0.0.0.0"

// ResolveLocalListenAddr returns an available TCP listen address for local development.
// It starts from preferredAddr, or DefaultHost:defaultPort when empty, and increments
// the port until one is free.
func ResolveLocalListenAddr(preferredAddr string, defaultPort int) (addr string, requestedPort int, err error) {
	host, startPort := ParseHostPort(preferredAddr, DefaultHost, defaultPort)
	port, err := NextAvailablePort(host, startPort)
	if err != nil {
		return "", startPort, err
	}
	return net.JoinHostPort(host, strconv.Itoa(port)), startPort, nil
}

// ResolveLocalHTTPAddr is an alias for ResolveLocalListenAddr.
func ResolveLocalHTTPAddr(preferredAddr string, defaultPort int) (addr string, requestedPort int, err error) {
	return ResolveLocalListenAddr(preferredAddr, defaultPort)
}

// ParseHostPort parses a listen address, falling back to defaults when addr is empty or invalid.
func ParseHostPort(addr, defaultHost string, defaultPort int) (host string, port int) {
	host = defaultHost
	port = defaultPort
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return host, port
	}
	h, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return host, port
	}
	if strings.TrimSpace(h) != "" {
		host = h
	}
	if n, err := strconv.Atoi(portStr); err == nil && n > 0 {
		port = n
	}
	return host, port
}

// NextAvailablePort returns the first available TCP port starting from startPort.
func NextAvailablePort(host string, startPort int) (int, error) {
	if startPort <= 0 {
		startPort = 1
	}
	for port := startPort; port <= 65535; port++ {
		if isPortAvailable(host, port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found from %d", startPort)
}

func isPortAvailable(host string, port int) bool {
	ln, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
