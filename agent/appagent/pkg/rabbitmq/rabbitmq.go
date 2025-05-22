package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zrik/agent/appagent/pkg/config"
	"github.com/zrik/agent/appagent/pkg/logger"
)

// Service represents a RabbitMQ service
type Service struct {
	config     *config.RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
	closed     chan struct{}
	tasks      chan Task
}

// Task represents a task to be processed
type Task struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
	Source  SourceType      `json:"source"`
}

// NewService creates a new RabbitMQ service
func NewService(cfg *config.RabbitMQConfig) *Service {
	return &Service{
		config: cfg,
		closed: make(chan struct{}),
		tasks:  make(chan Task, 100), // Buffer for 100 tasks
	}
}

// Connect establishes a connection to RabbitMQ
func (s *Service) Connect() error {
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

	// Set QoS
	err = s.channel.Qos(
		s.config.PrefetchCount, // prefetch count
		0,                      // prefetch size
		false,                  // global
	)
	if err != nil {
		s.channel.Close()
		s.connection.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
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

	// Declare a queue
	s.queue, err = s.channel.QueueDeclare(
		s.config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		s.channel.Close()
		s.connection.Close()
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Bind the queue to the exchange with routing keys
	for _, key := range s.config.RoutingKeys {
		err = s.channel.QueueBind(
			s.queue.Name,          // queue name
			key,                   // routing key
			s.config.ExchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			s.channel.Close()
			s.connection.Close()
			return fmt.Errorf("failed to bind a queue: %w", err)
		}
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
			time.Sleep(s.config.ReconnectInterval * time.Second)
			err := s.Connect()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to reconnect to RabbitMQ")
				continue
			}
			logger.Info().Msg("Successfully reconnected to RabbitMQ")

			// Start consuming again
			go s.startConsuming()
			return
		}
	}
}

// startConsuming starts consuming messages from the queue
func (s *Service) startConsuming() error {
	msgs, err := s.channel.Consume(
		s.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			var task Task
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				logger.Error().Err(err).Msg("Error parsing message")
				d.Nack(false, false) // Reject the message without requeuing
				continue
			}

			// Check if the topic is the priority topic
			if task.Topic == s.config.PriorityTopic {
				// Send to task channel with priority
				select {
				case s.tasks <- task:
					d.Ack(false) // Acknowledge the message
				case <-s.closed:
					return
				}
			} else {
				// For non-priority topics, we'll still process them but with lower priority
				// This is handled by the task processor
				select {
				case s.tasks <- task:
					d.Ack(false) // Acknowledge the message
				case <-s.closed:
					return
				}
			}
		}
	}()

	return nil
}

// Start starts the RabbitMQ service
func (s *Service) Start() error {
	err := s.Connect()
	if err != nil {
		return err
	}

	return s.startConsuming()
}

// GetTasks returns the tasks channel
func (s *Service) GetTasks() <-chan Task {
	return s.tasks
}

// Close closes the RabbitMQ service
func (s *Service) Close() error {
	close(s.closed)

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
