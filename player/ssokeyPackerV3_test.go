package player

import (
	"fmt"
	"testing"

	minitoken "github.com/apexkit/gamekit/player/mini_token"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestEncodedAndDecodedSSOKeyV3(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	payload := &minitoken.TokenPayload{
		PlayerId:  "11254455",
		AppId:     "11254aa5dadsdadfasdfasdfasdfasdfs5",
		GameBrand: "jili",
		GameId:    "1125",
	}

	encoded, err := EncodedSSOKeyV3(rdb, payload)
	if err != nil {
		t.Errorf("EncodedSSOKeyV3 should not return error: %v", err)
	}
	fmt.Printf("t: %v, encoded: %s\n", t, encoded)

	decoded, err := DecodedSSOKeyV3(rdb, encoded)
	if err != nil {
		t.Errorf("DecodedSSOKeyV3 should not return error: %v", err)
	}
	fmt.Printf("t: %v, decoded: %v\n", t, decoded)
}
