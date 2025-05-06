package http

import (
	"github.com/zrik/agent/appagent/pkg/config"
)

// Service represents the HTTP service
type Service struct {
	client        *Client
	taskResultSvc *TaskResultService
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

	return &Service{
		client:        client,
		taskResultSvc: taskResultSvc,
		config:        cfg,
	}
}

// GetTaskResultService returns the task result service
func (s *Service) GetTaskResultService() *TaskResultService {
	return s.taskResultSvc
}

// IsReportingEnabled returns whether result reporting is enabled
func (s *Service) IsReportingEnabled() bool {
	return s.config.ReportResults
}
