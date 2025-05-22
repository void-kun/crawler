package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"cct/config"
	"cct/pkg/logger"
)

// SourceType represents the source of a task
type SourceType string

// TaskType represents the type of a task
type TaskType string

// Known source types
const (
	SourceTypeSangTacViet SourceType = "sangtacviet"
	SourceTypeWikiDich    SourceType = "wikidich"
	SourceTypeMetruyenchu SourceType = "metruyenchu"
)

// Known task types
const (
	TaskTypeBook    TaskType = "book"
	TaskTypeChapter TaskType = "chapter"
	TaskTypeSession TaskType = "session"
)

// Task represents a task to be processed
type Task struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
	Source  SourceType      `json:"source"`
}

// BookTask represents a book task
type BookTask struct {
	BookURL string `json:"book_url"`
}

// ChapterTask represents a chapter task
type ChapterTask struct {
	ChapterURL string `json:"chapter_url"`
}

// SessionTask represents a session task
type SessionTask struct {
	URL string `json:"url"`
}

// Service represents a RabbitMQ service
type Service struct {
	config     *config.RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	closed     chan struct{}
	mu         sync.Mutex
}

// NewService creates a new RabbitMQ service
func NewService(cfg *config.RabbitMQConfig) *Service {
	return &Service{
		config: cfg,
		closed: make(chan struct{}),
	}
}

// Connect establishes a connection to RabbitMQ
func (s *Service) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	// Connect to RabbitMQ server
	s.connection, err = amqp.Dial(s.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	s.channel, err = s.connection.Channel()
	if err != nil {
		s.connection.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare an exchange
	err = s.channel.ExchangeDeclare(
		s.config.ExchangeName, // name
		s.config.ExchangeType, // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		s.channel.Close()
		s.connection.Close()
		return fmt.Errorf("failed to declare an exchange: %w", err)
	}

	// Set up connection close notifier
	closeChan := make(chan *amqp.Error)
	s.channel.NotifyClose(closeChan)

	// Handle connection close
	go func() {
		<-closeChan
		logger.Info().Msg("RabbitMQ connection closed, attempting to reconnect...")
		s.reconnect()
	}()

	return nil
}

// reconnect attempts to reconnect to RabbitMQ
func (s *Service) reconnect() {
	for {
		select {
		case <-s.closed:
			return
		default:
			time.Sleep(s.config.ReconnectInterval)
			err := s.Connect()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to reconnect to RabbitMQ")
				continue
			}
			logger.Info().Msg("Successfully reconnected to RabbitMQ")
			return
		}
	}
}

// Close closes the RabbitMQ service
func (s *Service) Close() error {
	close(s.closed)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.channel != nil {
		s.channel.Close()
	}

	if s.connection != nil {
		return s.connection.Close()
	}

	return nil
}

// PublishTask publishes a task to RabbitMQ
func (s *Service) PublishTask(ctx context.Context, task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.channel == nil {
		return errors.New("channel is nil, connection may be closed")
	}

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	return s.channel.PublishWithContext(
		ctx,
		s.config.ExchangeName, // exchange
		task.Topic,            // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

// GetTopicFromTaskTypeAndSource returns the topic for a task type and source
func GetTopicFromTaskTypeAndSource(taskType TaskType, source SourceType) string {
	return fmt.Sprintf("crawl.%s.%s", source, taskType)
}

// CreateBookTask creates a new book task
func CreateBookTask(source SourceType, bookURL string) Task {
	payload, _ := json.Marshal(BookTask{
		BookURL: bookURL,
	})

	return Task{
		Topic:   GetTopicFromTaskTypeAndSource(TaskTypeBook, source),
		Payload: payload,
		Source:  source,
	}
}

// CreateChapterTask creates a new chapter task
func CreateChapterTask(source SourceType, chapterURL string) Task {
	payload, _ := json.Marshal(ChapterTask{
		ChapterURL: chapterURL,
	})

	return Task{
		Topic:   GetTopicFromTaskTypeAndSource(TaskTypeChapter, source),
		Payload: payload,
		Source:  source,
	}
}

// CreateSessionTask creates a new session task
func CreateSessionTask(source SourceType, url string) Task {
	payload, _ := json.Marshal(SessionTask{
		URL: url,
	})

	return Task{
		Topic:   GetTopicFromTaskTypeAndSource(TaskTypeSession, source),
		Payload: payload,
		Source:  source,
	}
}
