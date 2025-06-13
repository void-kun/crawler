package main

import (
	"context"
	"flag"
	"log" // Standard log for initialization only
	"os"
	"time"

	"github.com/zrik/agent/appagent/internal/source/stv"
	"github.com/zrik/agent/appagent/pkg/config"
	"github.com/zrik/agent/appagent/pkg/logger"
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

	// Initialize logger
	loggerConfig := &logger.Config{
		Level:      cfg.Logger.Level,
		Output:     cfg.Logger.Output,
		FilePath:   cfg.Logger.FilePath,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}

	if err := logger.Init(loggerConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info().Msg("Starting RabbitMQ worker...")

	// Create application service
	service := rabbitmq.NewAppService(cfg)

	// Loop heartbeat to control API
	go func() {
		agentID := service.GetHTTPService().GetAgent().ID.String()
		ctx := context.Background()
		for {
			if err := service.GetHTTPService().GetAgentService().Heartbeat(ctx, agentID); err != nil {
				logger.Error().Err(err).Msg("Error sending heartbeat")
			}
			time.Sleep(cfg.ControlAPI.AgentHeartbeatInterval * time.Second)
		}
	}()

	// Register source clients
	websites, err := service.GetHTTPService().GetWebsiteService().GetWebsites(context.Background())
	if err != nil {
		logger.Fatal().Err(err).Msg("Error getting websites")
	}

	for _, website := range websites {
		switch website.ScriptName {
		case string(rabbitmq.SourceTypeSangTacViet):
			stvClient := stv.New(website.Username, website.Password, website.URL)
			service.RegisterSourceClient(rabbitmq.SourceTypeSangTacViet, stvClient)
		case string(rabbitmq.SourceTypeMetruyenchu):
			// metruyenchuClient := &metruyenchu.Metruyenchu{...}
			// service.RegisterSourceClient(rabbitmq.SourceTypeMetruyenchu, metruyenchuClient)
		case string(rabbitmq.SourceTypeWikiDich):
			// wikiDichClient := &wikidich.WikiDich{...}
			// service.RegisterSourceClient(rabbitmq.SourceTypeWikiDich, wikiDichClient)
		}
	}

	// Start the service
	if err := service.Start(); err != nil {
		logger.Error().Err(err).Msg("Error starting service")
		os.Exit(1)
	}
}
