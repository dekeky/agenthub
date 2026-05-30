package main

import (
	"flag"
	"log"

	"github.com/agenthub/internal/config"
	"github.com/agenthub/internal/hub"
	"github.com/agenthub/internal/router"
)

func main() {
	configPath := flag.String("config", "", "path to agenthub-server.toml (default: ./agenthub-server.toml)")
	flag.Parse()

	cfg, err := config.LoadServer(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	store, err := router.MustInitStore(cfg)
	if err != nil {
		log.Fatalf("failed to init store: %v", err)
	}

	svc := hub.NewService(store)
	app := router.New(cfg, svc)
	if err := app.Init(); err != nil {
		log.Fatalf("failed to init router: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
