package listen

import "testing"

func TestNormalizeLocal(t *testing.T) {
	addrs := Addresses{HTTP: "0.0.0.0:5002", GRPC: "0.0.0.0:9000"}
	t.Setenv("IS_LOCAL", "true")
	Normalize(&addrs, WithServiceName("test"))
	if addrs.HTTP == "" || addrs.GRPC == "" {
		t.Fatalf("addrs not resolved: %#v", addrs)
	}
}

func TestNormalizeProd(t *testing.T) {
	t.Setenv("IS_LOCAL", "")
	addrs := Addresses{HTTP: "0.0.0.0:5002", GRPC: "0.0.0.0:9000"}
	Normalize(&addrs)
	if addrs.HTTP != ProdHTTPAddr || addrs.GRPC != ProdGRPCAddr {
		t.Fatalf("got %#v", addrs)
	}
	if PprofListenAddr() != ProdPprofAddr {
		t.Fatalf("pprof=%q", PprofListenAddr())
	}
}

func TestPprofLocal(t *testing.T) {
	t.Setenv("IS_LOCAL", "true")
	addr := PprofListenAddr(WithServiceName("test"))
	host, port := ParseHostPort(addr, DefaultHost, DefaultLocalPprofPort)
	if host != DefaultHost || port < DefaultLocalPprofPort {
		t.Fatalf("pprof=%q host=%s port=%d", addr, host, port)
	}
}
