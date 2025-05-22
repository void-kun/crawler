package http

import (
	"context"
	"fmt"
	"io"
	"time"
)

// TaskResultStatus represents the status of a task result
type TaskResultStatus string

const (
	// TaskResultStatusSuccess indicates a successful task
	TaskResultStatusSuccess TaskResultStatus = "success"
	// TaskResultStatusError indicates a failed task
	TaskResultStatusError TaskResultStatus = "error"
)

// TaskType represents the type of task
type TaskType string

// SourceType represents the source of the task
type SourceType string

const (
	// TaskTypeBook represents a book crawling task
	TaskTypeBook TaskType = "book"
	// TaskTypeChapter represents a chapter crawling task
	TaskTypeChapter TaskType = "chapter"
	// TaskTypeSession represents a session extraction task
	TaskTypeSession TaskType = "session"
)

const (
	// SourceTypeSangTacViet represents the SangTacViet source
	SourceTypeSangTacViet SourceType = "sangtacviet"
	// SourceTypeWikiDich represents the WikiDich source
	SourceTypeWikiDich SourceType = "wikidich"
	// SourceTypeMetruyenchu represents the Metruyenchu source
	SourceTypeMetruyenchu SourceType = "metruyenchu"
)

// TaskResult represents the result of a task
type TaskResult struct {
	TaskID      string           `json:"task_id"`
	TaskType    TaskType         `json:"task_type"`
	Source      SourceType       `json:"source"`
	Status      TaskResultStatus `json:"status"`
	Message     string           `json:"message"`
	Data        interface{}      `json:"data,omitempty"`
	URL         string           `json:"url"`
	CompletedAt time.Time        `json:"completed_at"`
}

// TaskService handles task result reporting
type TaskService struct {
	client *Client
}

// NewTaskResultService creates a new task result service
func NewTaskResultService(client *Client) *TaskService {
	return &TaskService{
		client: client,
	}
}

// ReportTaskResult reports a task result to the control API
func (s *TaskService) ReportTaskResult(ctx context.Context, result *TaskResult) error {
	// Set the completion time if not already set
	if result.CompletedAt.IsZero() {
		result.CompletedAt = time.Now()
	}

	// Make the request
	resp, err := s.client.Post(ctx, "/api/task-results", result)
	if err != nil {
		return fmt.Errorf("failed to report task result: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to report task result: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ReportTaskSuccess reports a successful task result
func (s *TaskService) ReportTaskSuccess(ctx context.Context, taskID string, taskType TaskType, source SourceType, url string, data interface{}) error {
	result := &TaskResult{
		TaskID:      taskID,
		TaskType:    taskType,
		Source:      source,
		Status:      TaskResultStatusSuccess,
		Message:     "Task completed successfully",
		Data:        data,
		URL:         url,
		CompletedAt: time.Now(),
	}

	return s.ReportTaskResult(ctx, result)
}

// ReportTaskError reports a failed task result
func (s *TaskService) ReportTaskError(ctx context.Context, taskID string, taskType TaskType, source SourceType, url string, err error) error {
	result := &TaskResult{
		TaskID:      taskID,
		TaskType:    taskType,
		Source:      source,
		Status:      TaskResultStatusError,
		Message:     err.Error(),
		URL:         url,
		CompletedAt: time.Now(),
	}

	return s.ReportTaskResult(ctx, result)
}

func (s *TaskService) CreateTask(ctx context.Context)

// GenerateTaskID generates a task ID from a task URL and type
func GenerateTaskID(taskType TaskType, source SourceType, url string) string {
	return fmt.Sprintf("%s-%s-%s-%d", string(source), string(taskType), url, time.Now().Unix())
}
