package bootstrap

import (
	"flag"
	"fmt"
	"os"

	common_http "github.com/apexkit/gamekit/infra/transport/http"
	"github.com/bwmarrin/snowflake"
)

// Init runs shared process setup and registers the -conf flag.
// Returns a stable service instance id: hostname + snowflake.
func Init(confPath *string) string {
	common_http.InitDefaultTransport()

	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(fmt.Sprintf("init snowflake node error: %v", err))
	}

	hostname, _ := os.Hostname()
	flag.StringVar(confPath, "conf", "./configs", "config path, eg: -conf config.yaml")
	return hostname + node.Generate().String()
}
