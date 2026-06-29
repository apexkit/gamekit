package resource

import (
	"testing"

	gkconf "github.com/apexkit/gamekit/conf"
)

func TestValidate_WithNats(t *testing.T) {
	opts := &Options{WithNats: true}
	bc := &gkconf.Bootstrap{
		Data: &gkconf.Data{
			Eventbus: &gkconf.Data_Eventbus{Type: "nats", Url: "nats://127.0.0.1:4222"},
		},
	}
	if err := validate(opts, bc); err != nil {
		t.Fatal(err)
	}
	bc.Data.Eventbus = nil
	if err := validate(opts, bc); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate_WithMysqlRedis(t *testing.T) {
	opts := &Options{WithMysql: true, WithRedis: true}
	bc := &gkconf.Bootstrap{
		Data: &gkconf.Data{
			Database: []*gkconf.Data_Database{{Name: "default", Type: "mysql"}},
			Redis:    &gkconf.Data_Redis{Host: "127.0.0.1", Port: 6379},
		},
	}
	if err := validate(opts, bc); err != nil {
		t.Fatal(err)
	}
}
