package main

import (
	"github.com/daariikk/MyHelp/services/account-service/internal/app"
	"github.com/daariikk/MyHelp/services/account-service/internal/config"
)

func main() {
	cfg := config.MustLoad()
	application := app.New(cfg)
	application.Run()
}
