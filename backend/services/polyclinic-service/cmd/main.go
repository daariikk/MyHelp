package main

import (
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/app"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/config"
)

func main() {
	cfg := config.MustLoad()
	application := app.New(cfg)
	application.Run()
}
