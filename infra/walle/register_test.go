package walle

import (
	"testing"

	"github.com/apexkit/gamekit/infra/listen"
)

func TestSkipConsulRegister(t *testing.T) {
	t.Setenv(listen.EnvLocal, "")
	t.Setenv(EnvGroup, "")
	if SkipConsulRegister() {
		t.Fatal("expected false when IS_LOCAL and GROUP are unset")
	}

	t.Setenv(listen.EnvLocal, "true")
	t.Setenv(EnvGroup, "")
	if SkipConsulRegister() {
		t.Fatal("expected false when only IS_LOCAL is set")
	}

	t.Setenv(listen.EnvLocal, "")
	t.Setenv(EnvGroup, "test")
	if SkipConsulRegister() {
		t.Fatal("expected false when only GROUP is set")
	}

	t.Setenv(listen.EnvLocal, "true")
	t.Setenv(EnvGroup, "test")
	if !SkipConsulRegister() {
		t.Fatal("expected true when IS_LOCAL and GROUP are set")
	}
}
