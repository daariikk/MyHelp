package main

import (
	"github.com/daariikk/MyHelp/services/appointment-service/internal/app"
	"github.com/daariikk/MyHelp/services/appointment-service/internal/config"
)

func main() {
	cfg := config.MustLoad()
	application := app.New(cfg)
	application.Run()
}
