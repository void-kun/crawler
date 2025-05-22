package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"cct/models"

	"github.com/google/uuid"
)

// GetAgents handles GET /agents?active_only={true|false}&ip_address={ip_address}&name={name}
func GetAgents(w http.ResponseWriter, r *http.Request) {
	// Check if active_only query parameter is provided
	activeOnly := r.URL.Query().Get("active_only") == "true"
	ipAddress := r.URL.Query().Get("ip_address")
	name := r.URL.Query().Get("name")

	var agents []models.Agent
	var err error

	if activeOnly {
		// Get only active agents
		agents, err = models.GetActiveAgents(&ipAddress, &name)
	} else {
		// Get all agents
		agents, err = models.GetAgents(&ipAddress, &name)
	}

	if err != nil {
		http.Error(w, "Failed to get agents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// GetAgent handles GET /agents/{id}
func GetAgent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	agent, err := models.GetAgent(id)
	if err != nil {
		http.Error(w, "Failed to get agent: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// CreateAgent handles POST /agents
func CreateAgent(w http.ResponseWriter, r *http.Request) {
	var agent models.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if !agent.IsActive {
		agent.IsActive = true
	}

	if err := models.CreateAgent(&agent); err != nil {
		http.Error(w, "Failed to create agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

// UpdateAgent handles PUT /agents/{id}
func UpdateAgent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	var agent models.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	agent.ID = id

	if err := models.UpdateAgent(&agent); err != nil {
		http.Error(w, "Failed to update agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// HeartbeatAgent handles POST /agents/{id}/heartbeat
func HeartbeatAgent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Get IP address from request
	ipAddress := r.RemoteAddr

	// Update heartbeat
	if err := models.UpdateAgentHeartbeat(id, ipAddress); err != nil {
		http.Error(w, "Failed to update agent heartbeat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeactivateInactiveAgents handles POST /agents/deactivate-inactive
func DeactivateInactiveAgents(w http.ResponseWriter, r *http.Request) {
	// Parse request body for inactive duration
	var requestBody struct {
		InactiveDurationSeconds int `json:"inactive_duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Default to 5 minutes if not provided
	if requestBody.InactiveDurationSeconds <= 0 {
		requestBody.InactiveDurationSeconds = 300 // 5 minutes
	}

	// Convert to duration
	inactiveDuration := time.Duration(requestBody.InactiveDurationSeconds) * time.Second

	// Deactivate inactive agents
	count, err := models.DeactivateInactiveAgents(inactiveDuration)
	if err != nil {
		http.Error(w, "Failed to deactivate inactive agents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the number of deactivated agents
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"deactivated_count": count})
}

// DeleteAgent handles DELETE /agents/{id}
func DeleteAgent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteAgent(id); err != nil {
		http.Error(w, "Failed to delete agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
