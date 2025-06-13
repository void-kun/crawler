package http

import (
	"context"
	"time"

	"github.com/zrik/agent/appagent/pkg/config"
	"github.com/zrik/agent/appagent/pkg/logger"
)

type IService interface {
	GetTaskService() ITaskService
	GetAgentService() IAgentService
	GetWebsiteService() IWebsiteService
	IsReportingEnabled() bool
	GetAgent() *Agent
}

// Service represents the HTTP service
type Service struct {
	agent      *Agent
	config     *config.ControlAPIConfig
	taskSvc    ITaskService
	agentSvc   IAgentService
	websiteSvc IWebsiteService
}

// NewService creates a new HTTP service
func NewService(cfg *config.ControlAPIConfig) IService {
	// Create the HTTP client
	client := NewClient(cfg.BaseURL, cfg.Timeout*time.Second)

	// client.SetHeader("Content-Type", "application/json")
	client.SetHeader("Accept", "*/*")
	// Set the API key if provided
	if cfg.APIKey == "" {
		logger.Fatal().Msg("API key not provided")
	}
	client.SetHeader("Api-Key", cfg.APIKey)

	// Create the task result service
	taskSvc := NewTaskResultService(client)
	agentSvc := NewAgentService(client)
	websiteSvc := NewWebsiteService(client)

	agent, err := agentSvc.GetAgent(context.Background(), cfg.IPAddress, cfg.AgentName)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to get agent")
	} else {
		logger.Info().Str("agent", agent.Name).Msg("Agent registered")
	}

	return &Service{
		config:     cfg,
		agent:      agent,
		taskSvc:    taskSvc,
		agentSvc:   agentSvc,
		websiteSvc: websiteSvc,
	}
}

// GetTaskResultService returns the task result service
func (s *Service) GetTaskService() ITaskService {
	return s.taskSvc
}

func (s *Service) GetWebsiteService() IWebsiteService {
	return s.websiteSvc
}

func (s *Service) GetAgentService() IAgentService {
	return s.agentSvc
}

func (s *Service) GetAgent() *Agent {
	return s.agent
}

// IsReportingEnabled returns whether result reporting is enabled
func (s *Service) IsReportingEnabled() bool {
	return s.config.ReportResults
}
