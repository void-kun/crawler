package rabbitmq

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zrik/agent/appagent/internal/source"
	"github.com/zrik/agent/appagent/pkg/config"
	http "github.com/zrik/agent/appagent/pkg/http"
	"github.com/zrik/agent/appagent/pkg/spider"
)

// AppService represents the main application service
type AppService struct {
	config        *config.Config
	rabbitMQ      *Service
	spider        *spider.HeadSpider
	sourceClients map[SourceType]source.WebSource
	processor     *Processor
	httpService   *http.Service
}

// NewAppService creates a new application service
func NewAppService(cfg *config.Config) *AppService {
	// Create RabbitMQ service
	rabbitMQ := NewService(&cfg.RabbitMQ)

	// Create HTTP service if control API is configured
	var httpService *http.Service
	if cfg.ControlAPI.BaseURL != "" {
		httpService = http.NewService(&cfg.ControlAPI)
		log.Printf("Control API enabled at %s", cfg.ControlAPI.BaseURL)
	}

	// Create spider
	spiderInstance := spider.NewHeadSpider(true, cfg)

	// Create processor
	processor := NewProcessor(rabbitMQ, cfg, spiderInstance)

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
	s.sourceClients[sourceType] = client
	s.processor.RegisterSourceClient(sourceType, client)
}

// Start starts the application service
func (s *AppService) Start() error {
	log.Println("Starting application service...")
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
	log.Println("Received shutdown signal, gracefully shutting down...")

	// Stop processor
	s.processor.Stop()

	// Close RabbitMQ connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.rabbitMQ.Close(); err != nil {
		log.Printf("Error closing RabbitMQ connection: %v", err)
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
		log.Printf("Error closing RabbitMQ connection: %v", err)
	}
}
