// Backgammon v2
package main

import (
	"os"

	"github.com/therealtunge/backgammon/client"
	"github.com/therealtunge/backgammon/common/logging"
	"github.com/therealtunge/backgammon/server"
)

func main() {
	if os.Args[1] == "client" {
		logging.Warn("loading client")
		client.Main()
	} else {
		logging.Warn("loading server")
		server.Main()
	}
}
