package rabbitmq

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zrik/agent/appagent/internal/source"
	"github.com/zrik/agent/appagent/pkg/config"
	http "github.com/zrik/agent/appagent/pkg/http"
	"github.com/zrik/agent/appagent/pkg/logger"
	"github.com/zrik/agent/appagent/pkg/spider"
)

// AppService represents the main application service
type AppService struct {
	config        *config.Config
	rabbitMQ      *Service
	spider        *spider.HeadSpider
	sourceClients map[SourceType]source.WebSource
	processor     *Processor
	httpService   http.IService
}

// NewAppService creates a new application service
func NewAppService(cfg *config.Config) *AppService {
	// Create RabbitMQ service
	rabbitMQ := NewService(&cfg.RabbitMQ)

	// Create HTTP service if control API is configured
	var httpService http.IService
	if cfg.ControlAPI.BaseURL != "" {
		httpService = http.NewService(&cfg.ControlAPI)
		logger.Info().Str("baseURL", cfg.ControlAPI.BaseURL).Msg("Control API enabled")
	}

	// Create spider
	spiderInstance := spider.NewHeadSpider(true, cfg)
	_, err := spiderInstance.CreatePage()
	if err != nil {
		logger.Error().Err(err).Msg("Error creating page")
	}

	// Load session data
	if err := spiderInstance.LoadSessionDataFromJSON(); err != nil {
		logger.Error().Err(err).Msg("Error loading session data")
	}

	// Create processor
	processor := NewProcessor(rabbitMQ, cfg, spiderInstance, httpService)

	// Register default task processors
	processor.RegisterDefaultTaskProcessors()

	return &AppService{
		config:        cfg,
		rabbitMQ:      rabbitMQ,
		spider:        spiderInstance,
		sourceClients: make(map[SourceType]source.WebSource),
		processor:     processor,
		httpService:   httpService,
	}
}

// RegisterSourceClient registers a source client
func (s *AppService) RegisterSourceClient(sourceType SourceType, client source.WebSource) {
	logger.Info().Str("source", string(sourceType)).Msg("Registering source client")
	s.sourceClients[sourceType] = client
	s.processor.RegisterSourceClient(sourceType, client)
}

// Start starts the application service
func (s *AppService) Start() error {
	logger.Info().Msg("Starting application service...")
	// Start RabbitMQ service
	if err := s.rabbitMQ.Start(); err != nil {
		return err
	}

	// Start task processor
	s.processor.Start()
	s.processor.ProcessTasks()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	logger.Info().Msg("Received shutdown signal, gracefully shutting down...")

	// Stop processor
	s.processor.Stop()

	// Close RabbitMQ connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.rabbitMQ.Close(); err != nil {
		logger.Error().Err(err).Msg("Error closing RabbitMQ connection")
	}

	// Wait for context to be done
	<-ctx.Done()

	return nil
}

// Stop stops the application service
func (s *AppService) Stop() {
	// Stop processor
	s.processor.Stop()

	// Close RabbitMQ connection
	if err := s.rabbitMQ.Close(); err != nil {
		logger.Error().Err(err).Msg("Error closing RabbitMQ connection")
	}
}

func (s *AppService) GetHTTPService() http.IService {
	return s.httpService
}
