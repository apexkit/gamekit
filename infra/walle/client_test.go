package walle

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_HTTPErrorIncludesURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.GetGameGroup(context.Background(), "prod_games")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "/openapi/game/group") {
		t.Fatalf("error should include request path: %q", msg)
	}
	if !strings.Contains(msg, "HTTP 404") {
		t.Fatalf("error should include status code: %q", msg)
	}
	if !strings.Contains(msg, `group "prod_games"`) {
		t.Fatalf("error should include group name: %q", msg)
	}
}

func TestClient_GetGameGroup(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openapi/game/group" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("group"); got != "prod_games" {
			t.Fatalf("unexpected group query: %q", got)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Fatalf("unexpected auth: %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status":"ok",
			"message":"success",
			"data":[{
				"group_name":"prod_games",
				"log_level":"info",
				"mysql_config":{"internal_host":"10.0.0.5","port":3306,"connection_type":1,"account":"u","password":"p"},
				"redis_config":{"internal_endpoint":"127.0.0.1:6379","connection_type":1,"auth":"secret","db":1},
				"consul_config":{"internal_endpoint":"http://consul:8500","connection_type":1},
				"s3_config":{"access_key":"ak","secret_key":"sk","region":"ap-southeast-1","bucket":"wxgame"},
				"nats_config":{"id":3,"name":"nats-生产","internal_endpoint":"nats://127.0.0.1:4222","external_endpoint":"nats://nats.example.com:4222","connection_type":1}
			}]
		}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	group, err := client.GetGameGroup(context.Background(), "prod_games")
	if err != nil {
		t.Fatalf("GetGameGroup: %v", err)
	}
	if group == nil || group.GroupName != "prod_games" {
		t.Fatalf("unexpected group: %#v", group)
	}
	if group.NatsConfig == nil || group.NatsConfig.ID != 3 || group.NatsConfig.Name != "nats-生产" {
		t.Fatalf("unexpected nats_config: %#v", group.NatsConfig)
	}
}

func TestParseGroupName(t *testing.T) {
	name, err := ParseGroupName("  Prod_Games ")
	if err != nil {
		t.Fatal(err)
	}
	if name != "prod_games" {
		t.Fatalf("got %q", name)
	}

	_, err = ParseGroupName("a,b")
	if err == nil {
		t.Fatal("expected error for multiple groups")
	}
}

func TestMySQLDSN(t *testing.T) {
	t.Setenv("IS_LOCAL", "")

	dsn, err := MySQLDSN(&MySQLConfig{
		InternalHost: "10.0.0.5",
		ExternalHost: "217.15.162.172",
		Port:         3306,
		Account:      "game_user",
		Password:     "pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "game_user:pass@tcp(10.0.0.5:3306)/game?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != want {
		t.Fatalf("got %q want %q", dsn, want)
	}

	t.Setenv("IS_LOCAL", "true")
	dsn, err = MySQLDSN(&MySQLConfig{
		InternalHost: "10.0.0.5",
		ExternalHost: "217.15.162.172",
		Port:         33306,
		Account:      "game_user",
		Password:     "pass",
		Database:     "prod_game",
	})
	if err != nil {
		t.Fatal(err)
	}
	want = "game_user:pass@tcp(217.15.162.172:33306)/prod_game?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != want {
		t.Fatalf("got %q want %q", dsn, want)
	}
}

func TestConsulEndpoint(t *testing.T) {
	t.Setenv("IS_LOCAL", "")

	addr, scheme, err := ConsulEndpoint(&ConsulConfig{
		InternalEndpoint: "http://consul:8500",
		ExternalEndpoint: "http://consul.example.com:8500",
	})
	if err != nil {
		t.Fatal(err)
	}
	if addr != "consul:8500" || scheme != "http" {
		t.Fatalf("got %q %q", addr, scheme)
	}

	t.Setenv("IS_LOCAL", "true")
	addr, scheme, err = ConsulEndpoint(&ConsulConfig{
		InternalEndpoint: "http://consul:8500",
		ExternalEndpoint: "http://consul.example.com:8500",
	})
	if err != nil {
		t.Fatal(err)
	}
	if addr != "consul.example.com:8500" || scheme != "http" {
		t.Fatalf("got %q %q", addr, scheme)
	}
}

func TestNatsEndpoint(t *testing.T) {
	t.Setenv("IS_LOCAL", "")

	url, err := NatsEndpoint(&NatsConfig{
		InternalEndpoint: "nats://127.0.0.1:4222",
		ExternalEndpoint: "nats://nats.example.com:4222",
	})
	if err != nil {
		t.Fatal(err)
	}
	if url != "nats://127.0.0.1:4222" {
		t.Fatalf("got %q", url)
	}

	t.Setenv("IS_LOCAL", "true")
	url, err = NatsEndpoint(&NatsConfig{
		InternalEndpoint: "nats://127.0.0.1:4222",
		ExternalEndpoint: "nats://nats.example.com:4222",
	})
	if err != nil {
		t.Fatal(err)
	}
	if url != "nats://nats.example.com:4222" {
		t.Fatalf("got %q", url)
	}

	t.Setenv("IS_LOCAL", "")
	url, err = NatsEndpoint(&NatsConfig{InternalEndpoint: "127.0.0.1:4222"})
	if err != nil {
		t.Fatal(err)
	}
	if url != "nats://127.0.0.1:4222" {
		t.Fatalf("got %q", url)
	}
}
