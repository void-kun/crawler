package rabbitmq

import (
	"encoding/json"
	"fmt"
	"strings"
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
	SourceTypeSangTacViet SourceType = "sangtacviet"
	SourceTypeWikiDich    SourceType = "wikidich"
	SourceTypeMetruyenchu SourceType = "metruyenchu"
)

// TopicPrefix is the prefix for all topics
const TopicPrefix = "crawl."

// GetTopicFromTaskTypeAndSource returns the full topic name for a task type and source
func GetTopicFromTaskTypeAndSource(taskType TaskType, source SourceType) string {
	return TopicPrefix + string(source) + "." + string(taskType)
}

// ParseTopicInfo parses a topic string to extract source and task type
func ParseTopicInfo(topic string) (SourceType, TaskType, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 3 || parts[0] != "crawl" {
		return "", "", fmt.Errorf("invalid topic format: %s", topic)
	}

	return SourceType(parts[1]), TaskType(parts[2]), nil
}

// BookTask represents a task to crawl a book
type BookTask struct {
	BookURL string `json:"book_url"`
}

// ChapterTask represents a task to crawl a chapter
type ChapterTask struct {
	ChapterURL string `json:"chapter_url"`
}

// SessionTask represents a task to extract a session
type SessionTask struct {
	URL string `json:"url"`
}

// ParseTask parses a task based on its topic
func ParseTask(task Task) (interface{}, error) {
	// Parse the topic to get source and task type
	_, taskType, err := ParseTopicInfo(task.Topic)
	if err != nil {
		return nil, err
	}

	switch taskType {
	case TaskTypeBook:
		var bookTask BookTask
		err := json.Unmarshal(task.Payload, &bookTask)
		if err != nil {
			return nil, fmt.Errorf("failed to parse book task: %w", err)
		}
		return bookTask, nil
	case TaskTypeChapter:
		var chapterTask ChapterTask
		err := json.Unmarshal(task.Payload, &chapterTask)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chapter task: %w", err)
		}
		return chapterTask, nil
	case TaskTypeSession:
		var sessionTask SessionTask
		err := json.Unmarshal(task.Payload, &sessionTask)
		if err != nil {
			return nil, fmt.Errorf("failed to parse session task: %w", err)
		}
		return sessionTask, nil
	default:
		return nil, fmt.Errorf("unknown task type: %s", taskType)
	}
}

// CreateBookTask creates a new book task
func CreateBookTask(source SourceType, bookURL string) Task {
	payload, _ := json.Marshal(BookTask{
		BookURL: bookURL,
	})

	return Task{
		Topic:   GetTopicFromTaskTypeAndSource(TaskTypeBook, source),
		Payload: payload,
		Source:  source,
	}
}

// CreateChapterTask creates a new chapter task
func CreateChapterTask(source SourceType, chapterURL string) Task {
	payload, _ := json.Marshal(ChapterTask{
		ChapterURL: chapterURL,
	})

	return Task{
		Topic:   GetTopicFromTaskTypeAndSource(TaskTypeChapter, source),
		Payload: payload,
		Source:  source,
	}
}

// CreateSessionTask creates a new session task
func CreateSessionTask(source SourceType, url string) Task {
	payload, _ := json.Marshal(SessionTask{
		URL: url,
	})

	return Task{
		Topic:   GetTopicFromTaskTypeAndSource(TaskTypeSession, source),
		Payload: payload,
		Source:  source,
	}
}
