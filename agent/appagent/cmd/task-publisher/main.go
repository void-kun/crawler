package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/zrik/agent/appagent/pkg/config"
	"github.com/zrik/agent/appagent/pkg/rabbitmq"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	taskType := flag.String("type", "session", "Task type (book, chapter, session)")
	sourceType := flag.String("source", "sangtacviet", "Source type (sangtacviet, wikidich, metruyenchu)")
	url := flag.String("url", "", "URL for the task")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfigFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create RabbitMQ service
	service := rabbitmq.NewService(&cfg.RabbitMQ)
	if err := service.Connect(); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer service.Close()

	// Parse source type
	var source rabbitmq.SourceType
	switch *sourceType {
	case "sangtacviet":
		source = rabbitmq.SourceTypeSangTacViet
	case "wikidich":
		source = rabbitmq.SourceTypeWikiDich
	case "metruyenchu":
		source = rabbitmq.SourceTypeMetruyenchu
	default:
		log.Fatalf("Unknown source type: %s", *sourceType)
	}

	// Create task based on type
	var task rabbitmq.Task
	switch *taskType {
	case "book":
		if *url == "" {
			log.Fatalf("Book tasks require url, book-id, and book-host parameters")
		}
		task = rabbitmq.CreateBookTask(source, *url)
	case "chapter":
		if *url == "" {
			log.Fatalf("Chapter tasks require url, book-id, chapter-id, book-host, and book-sty parameters")
		}
		task = rabbitmq.CreateChapterTask(source, *url)
	case "session":
		if *url == "" {
			log.Fatalf("Session tasks require url parameter")
		}
		task = rabbitmq.CreateSessionTask(source, *url)
	default:
		log.Fatalf("Unknown task type: %s", *taskType)
	}

	// Publish task
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := service.PublishTask(ctx, task); err != nil {
		log.Fatalf("Failed to publish task: %v", err)
	}

	log.Printf("Successfully published %s task", *taskType)
}
