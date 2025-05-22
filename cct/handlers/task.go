package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cct/config"
	"cct/pkg/logger"
	"cct/pkg/rabbitmq"
)

var agentService *rabbitmq.AgentService

// InitRabbitMQService initializes the RabbitMQ service
func InitRabbitMQService(cfg *config.Config) error {
	var err error
	agentService, err = rabbitmq.NewAgentService(&cfg.RabbitMQ)
	if err != nil {
		return err
	}
	return nil
}

// CloseRabbitMQService closes the RabbitMQ service
func CloseRabbitMQService() error {
	if agentService != nil {
		return agentService.Close()
	}
	return nil
}

// PublishTaskRequest represents a request to publish a task
type PublishTaskRequest struct {
	Source     string `json:"source"`
	TaskType   string `json:"task_type"`
	URL        string `json:"url"`
	TimeoutSec int    `json:"timeout_sec,omitempty"`
}

type TaskResultStatus string

const (
	// TaskResultStatusSuccess indicates a successful task
	TaskResultStatusSuccess TaskResultStatus = "success"
	// TaskResultStatusError indicates a failed task
	TaskResultStatusError TaskResultStatus = "error"
)

type TaskResult struct {
	TaskID      string              `json:"task_id"`
	TaskType    rabbitmq.TaskType   `json:"task_type"`
	Source      rabbitmq.SourceType `json:"source"`
	Status      TaskResultStatus    `json:"status"`
	Message     string              `json:"message"`
	Data        interface{}         `json:"data,omitempty"`
	URL         string              `json:"url"`
	CompletedAt time.Time           `json:"completed_at"`
}

// PublishTask handles POST /tasks/publish
func PublishTask(w http.ResponseWriter, r *http.Request) {
	if agentService == nil {
		http.Error(w, "RabbitMQ service not initialized", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req PublishTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Source == "" {
		http.Error(w, "Source is required", http.StatusBadRequest)
		return
	}
	if req.TaskType == "" {
		http.Error(w, "Task type is required", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Convert source to SourceType
	var source rabbitmq.SourceType
	switch req.Source {
	case "sangtacviet":
		source = rabbitmq.SourceTypeSangTacViet
	case "wikidich":
		source = rabbitmq.SourceTypeWikiDich
	case "metruyenchu":
		source = rabbitmq.SourceTypeMetruyenchu
	default:
		http.Error(w, "Invalid source: "+req.Source, http.StatusBadRequest)
		return
	}

	// Set default timeout if not provided
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 30
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(req.TimeoutSec)*time.Second)
	defer cancel()

	// Publish task based on task type
	var err error
	switch req.TaskType {
	case "book":
		err = agentService.PublishBookTask(ctx, source, req.URL)
	case "chapter":
		err = agentService.PublishChapterTask(ctx, source, req.URL)
	case "session":
		err = agentService.PublishSessionTask(ctx, source, req.URL)
	default:
		http.Error(w, "Invalid task type: "+req.TaskType, http.StatusBadRequest)
		return
	}

	if err != nil {
		logger.Error().Err(err).Str("source", req.Source).Str("task_type", req.TaskType).Str("url", req.URL).Msg("Failed to publish task")
		http.Error(w, "Failed to publish task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Task published successfully",
	})
}

// ResultTask handles POST /tasks/result
func ResultTask(w http.ResponseWriter, r *http.Request) {
}
