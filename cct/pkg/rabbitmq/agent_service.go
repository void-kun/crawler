package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cct/config"
	"cct/models"
	"cct/pkg/logger"
)

// AgentService represents a service for sending messages to agents
type AgentService struct {
	rabbitmq      *Service
	config        *config.RabbitMQConfig
	mu            sync.Mutex
	activeAgents  []models.Agent
	lastFetchTime time.Time
	cacheDuration time.Duration
}

// NewAgentService creates a new agent service
func NewAgentService(cfg *config.RabbitMQConfig) (*AgentService, error) {
	// Create RabbitMQ service
	rabbitmq := NewService(cfg)
	if err := rabbitmq.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &AgentService{
		rabbitmq:      rabbitmq,
		config:        cfg,
		cacheDuration: 30 * time.Second, // Cache active agents for 30 seconds
	}, nil
}

// Close closes the agent service
func (s *AgentService) Close() error {
	return s.rabbitmq.Close()
}

// refreshActiveAgents refreshes the list of active agents
func (s *AgentService) refreshActiveAgents() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we need to refresh the cache
	if time.Since(s.lastFetchTime) < s.cacheDuration && len(s.activeAgents) > 0 {
		return nil
	}

	// Get active agents from the database
	agents, err := models.GetActiveAgents(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to get active agents: %w", err)
	}

	// Update the cache
	s.activeAgents = agents
	s.lastFetchTime = time.Now()

	logger.Info().Int("count", len(agents)).Msg("Refreshed active agents")
	return nil
}

// PublishTaskToActiveAgents publishes a task to all active agents
func (s *AgentService) PublishTaskToActiveAgents(ctx context.Context, task Task) error {
	// Refresh the list of active agents
	if err := s.refreshActiveAgents(); err != nil {
		return err
	}

	s.mu.Lock()
	agentCount := len(s.activeAgents)
	s.mu.Unlock()

	if agentCount == 0 {
		return errors.New("no active agents available")
	}

	// Publish the task
	if err := s.rabbitmq.PublishTask(ctx, task); err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	logger.Info().
		Str("topic", task.Topic).
		Str("source", string(task.Source)).
		Int("agent_count", agentCount).
		Msg("Published task to active agents")

	return nil
}

// PublishBookTask publishes a book task to active agents
func (s *AgentService) PublishBookTask(ctx context.Context, source SourceType, bookURL string) error {
	task := CreateBookTask(source, bookURL)
	return s.PublishTaskToActiveAgents(ctx, task)
}

// PublishChapterTask publishes a chapter task to active agents
func (s *AgentService) PublishChapterTask(ctx context.Context, source SourceType, chapterURL string) error {
	task := CreateChapterTask(source, chapterURL)
	return s.PublishTaskToActiveAgents(ctx, task)
}

// PublishSessionTask publishes a session task to active agents
func (s *AgentService) PublishSessionTask(ctx context.Context, source SourceType, url string) error {
	task := CreateSessionTask(source, url)
	return s.PublishTaskToActiveAgents(ctx, task)
}

// GetActiveAgentCount returns the number of active agents
func (s *AgentService) GetActiveAgentCount() (int, error) {
	if err := s.refreshActiveAgents(); err != nil {
		return 0, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.activeAgents), nil
}
