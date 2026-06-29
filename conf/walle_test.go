package conf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apexkit/gamekit/infra/walle"
)

func TestApplyByWalle_NatsConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status":"ok",
			"message":"success",
			"data":[{
				"group_name":"prod_games",
				"nats_config":{
					"id":3,
					"name":"nats-生产",
					"internal_endpoint":"nats://127.0.0.1:4222",
					"external_endpoint":"nats://nats.example.com:4222",
					"connection_type":1
				}
			}]
		}`))
	}))
	defer srv.Close()

	t.Setenv(walle.EnvGroup, "prod_games")
	t.Setenv("WALLE_URL", srv.URL)
	t.Setenv("WALLE_TOKEN", "test-token")
	t.Setenv("IS_LOCAL", "")

	bc := &Bootstrap{Data: &Data{}}
	if err := ApplyByWalle(bc); err != nil {
		t.Fatalf("ApplyByWalle: %v", err)
	}
	if bc.Data.Eventbus == nil {
		t.Fatal("expected eventbus config")
	}
	if bc.Data.Eventbus.GetType() != "nats" {
		t.Fatalf("type=%q", bc.Data.Eventbus.GetType())
	}
	if bc.Data.Eventbus.GetUrl() != "nats://127.0.0.1:4222" {
		t.Fatalf("url=%q", bc.Data.Eventbus.GetUrl())
	}

	t.Setenv("IS_LOCAL", "true")
	bc = &Bootstrap{Data: &Data{}}
	if err := ApplyByWalle(bc); err != nil {
		t.Fatalf("ApplyByWalle local: %v", err)
	}
	if bc.Data.Eventbus.GetUrl() != "nats://nats.example.com:4222" {
		t.Fatalf("local url=%q", bc.Data.Eventbus.GetUrl())
	}
}
