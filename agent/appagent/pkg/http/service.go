package http

import (
	"context"

	"github.com/zrik/agent/appagent/pkg/config"
	"github.com/zrik/agent/appagent/pkg/logger"
)

// Service represents the HTTP service
type Service struct {
	agent         *Agent
	client        *Client
	taskResultSvc *TaskService
	agentSvc      *AgentService
	config        *config.ControlAPIConfig
}

// NewService creates a new HTTP service
func NewService(cfg *config.ControlAPIConfig) *Service {
	// Create the HTTP client
	client := NewClient(cfg.BaseURL, cfg.Timeout)

	// Set the API key if provided
	if cfg.APIKey != "" {
		client.SetHeader("X-API-Key", cfg.APIKey)
	}

	// Create the task result service
	taskResultSvc := NewTaskResultService(client)
	agentSvc := NewAgentService(client)

	agent, err := agentSvc.GetAgentByIpAddress(context.Background(), cfg.IPAddress, cfg.AgentName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get agent")
	} else {
		logger.Info().Str("agent", agent.Name).Msg("Agent registered")
	}

	return &Service{
		agent:         agent,
		client:        client,
		taskResultSvc: taskResultSvc,
		agentSvc:      agentSvc,
		config:        cfg,
	}
}

// GetTaskResultService returns the task result service
func (s *Service) GetTaskResultService() *TaskService {
	return s.taskResultSvc
}

// IsReportingEnabled returns whether result reporting is enabled
func (s *Service) IsReportingEnabled() bool {
	return s.config.ReportResults
}

func (s *Service) GetAgentService() *AgentService {
	return s.agentSvc
}

func (s *Service) GetAgent() *Agent {
	return s.agent
}
