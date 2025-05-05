package main

import (
	"flag"
	"log"
	"os"

	"github.com/zrik/agent/appagent/internal/source/stv"
	"github.com/zrik/agent/appagent/pkg/config"
	"github.com/zrik/agent/appagent/pkg/rabbitmq"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfigFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create application service
	service := rabbitmq.NewAppService(cfg)

	// Register source clients
	// SangTacViet source client
	stvClient := stv.New(cfg.Stv.Username, cfg.Stv.Password, cfg.Stv.Origin)
	service.RegisterSourceClient(rabbitmq.SourceTypeSangTacViet, stvClient)

	// Add more source clients here as needed
	// For example:
	// wikiDichClient := &wikidich.WikiDich{...}
	// service.RegisterSourceClient(rabbitmq.SourceTypeWikiDich, wikiDichClient)
	//
	// metruyenchuClient := &metruyenchu.Metruyenchu{...}
	// service.RegisterSourceClient(rabbitmq.SourceTypeMetruyenchu, metruyenchuClient)

	// Start the service
	if err := service.Start(); err != nil {
		log.Printf("Error starting service: %v", err)
		os.Exit(1)
	}
}
