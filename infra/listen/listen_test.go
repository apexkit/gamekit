package listen

import (
	"net"
	"strconv"
	"testing"
)

func TestParseHostPort(t *testing.T) {
	host, port := ParseHostPort("", DefaultHost, 5002)
	if host != DefaultHost || port != 5002 {
		t.Fatalf("default got %s:%d", host, port)
	}

	host, port = ParseHostPort("127.0.0.1:5003", DefaultHost, 5002)
	if host != "127.0.0.1" || port != 5003 {
		t.Fatalf("custom got %s:%d", host, port)
	}
}

func TestNextAvailablePort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	_, occupiedPortStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	occupiedPort, err := strconv.Atoi(occupiedPortStr)
	if err != nil {
		t.Fatal(err)
	}

	port, err := NextAvailablePort("127.0.0.1", occupiedPort)
	if err != nil {
		t.Fatal(err)
	}
	if port != occupiedPort+1 {
		t.Fatalf("got %d want %d", port, occupiedPort+1)
	}
}

func TestResolveLocalHTTPAddr(t *testing.T) {
	addr, requestedPort, err := ResolveLocalHTTPAddr("0.0.0.0:5002", 5002)
	if err != nil {
		t.Fatal(err)
	}
	if requestedPort != 5002 {
		t.Fatalf("requestedPort=%d", requestedPort)
	}
	host, port := ParseHostPort(addr, DefaultHost, 5002)
	if host != "0.0.0.0" || port < 5002 {
		t.Fatalf("resolved %s:%d", host, port)
	}
}
