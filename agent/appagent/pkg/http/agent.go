package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID            uuid.UUID    `json:"id"`
	Name          string       `json:"name"`
	IPAddress     string       `json:"ip_address"`
	LastHeartbeat sql.NullTime `json:"last_heartbeat"`
	IsActive      bool         `json:"is_active"`
	CreatedAt     time.Time    `json:"created_at"`
}

type AgentService struct {
	client *Client
}

// NewAgentService creates a new task result service
func NewAgentService(client *Client) *AgentService {
	return &AgentService{
		client: client,
	}
}

func (s *AgentService) GetAgentByIpAddress(ctx context.Context, ipAddress, name string) (*Agent, error) {
	// Make the request
	resp, err := s.client.Get(ctx, "/api/agents?only_active=true&ip_address="+ipAddress+"&name="+name)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get agent: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Decode the response
	var agents []Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, fmt.Errorf("failed to decode agent response: %w", err)
	}

	// Should only have one agent
	if len(agents) != 1 {
		return nil, fmt.Errorf("expected 1 agent, got %d", len(agents))
	}

	return &agents[0], nil
}

func (s *AgentService) Heartbeat(ctx context.Context, agentID string) error {
	// Make the request
	resp, err := s.client.Post(ctx, "/api/agents/"+agentID+"/heartbeat", nil)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send heartbeat: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *AgentService) IsActive(ctx context.Context, agentID string) (bool, error) {
	// Make the request
	resp, err := s.client.Get(ctx, "/api/agents/"+agentID)
	if err != nil {
		return false, fmt.Errorf("failed to check agent status: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to check agent status: status %d, body: %s", resp.StatusCode, string(body))
	}

	return true, nil
}
