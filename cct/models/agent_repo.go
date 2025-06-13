package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cct/utils"

	"github.com/google/uuid"
)

// GetAgent retrieves an agent by ID
func GetAgent(id uuid.UUID) (Agent, error) {
	var a Agent
	err := utils.DB.QueryRow(`
		SELECT id, name, ip_address, last_heartbeat, is_active, created_at
		FROM agents
		WHERE id = $1
	`, id).Scan(
		&a.ID, &a.Name, &a.IPAddress, &a.LastHeartbeat, &a.IsActive, &a.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Agent{}, fmt.Errorf("agent with ID %s not found", id)
		}
		return Agent{}, fmt.Errorf("failed to query agent: %w", err)
	}

	return a, nil
}

// GetAgents retrieves all agents
func GetAgents(isActive bool, ipAddress, name string) ([]Agent, error) {
	query := `
		SELECT id, name, ip_address, last_heartbeat, is_active, created_at
		FROM agents
	`

	params := []interface{}{}
	conditions := []string{}

	if isActive {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", len(conditions)+1))
		params = append(params, true)
	}

	if ipAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", len(conditions)+1))
		params = append(params, ipAddress)
	}

	if name != "" {
		conditions = append(conditions, fmt.Sprintf("name = $%d", len(conditions)+1))
		params = append(params, name)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := utils.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(
			&a.ID, &a.Name, &a.IPAddress, &a.LastHeartbeat, &a.IsActive, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan agent row: %w", err)
		}
		agents = append(agents, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating agent rows: %w", err)
	}

	return agents, nil
}

// CreateAgent creates a new agent in the database
func CreateAgent(a *Agent) error {
	// Check if agent with the same IP already exists
	var existingAgentID uuid.UUID
	err := utils.DB.QueryRow(`
		SELECT id, name, ip_address, last_heartbeat, is_active, created_at
		FROM agents
		WHERE ip_address = $1 AND name = $2
	`, a.IPAddress, a.Name).Scan(&existingAgentID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query existing agent: %w", err)
	}

	// Generate a new UUID if not provided
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	err = utils.DB.QueryRow(`
		INSERT INTO agents (id, name, ip_address, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, a.ID, a.Name, a.IPAddress, a.IsActive).Scan(&a.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

// UpdateAgent updates an existing agent
func UpdateAgent(a *Agent) error {
	_, err := utils.DB.Exec(`
		UPDATE agents
		SET name = $1, ip_address = $2, is_active = $3
		WHERE id = $4
	`, a.Name, a.IPAddress, a.IsActive, a.ID)
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	return nil
}

// UpdateAgentHeartbeat updates the last_heartbeat timestamp for an agent
func UpdateAgentHeartbeat(id uuid.UUID) error {
	_, err := utils.DB.Exec(`
		UPDATE agents
		SET last_heartbeat = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to update agent heartbeat: %w", err)
	}

	return nil
}

// DeactivateInactiveAgents marks agents as inactive if they haven't sent a heartbeat in the specified duration
func DeactivateInactiveAgents(inactiveDuration time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-inactiveDuration)

	result, err := utils.DB.Exec(`
		UPDATE agents
		SET is_active = false
		WHERE is_active = true AND (last_heartbeat IS NULL OR last_heartbeat < $1)
	`, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate inactive agents: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}

// DeleteAgent deletes an agent by ID
func DeleteAgent(id uuid.UUID) error {
	_, err := utils.DB.Exec("DELETE FROM agents WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	return nil
}
