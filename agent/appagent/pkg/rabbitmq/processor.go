package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/zrik/agent/appagent/internal/source"
	"github.com/zrik/agent/appagent/pkg/config"
	http "github.com/zrik/agent/appagent/pkg/http"
	"github.com/zrik/agent/appagent/pkg/logger"
	"github.com/zrik/agent/appagent/pkg/spider"
)

// SourceClientRegistry is a registry for source clients
type SourceClientRegistry map[SourceType]source.WebSource

// Processor represents a task processor
type Processor struct {
	service        *Service
	config         *config.Config
	spider         spider.TaskSpider
	sourceClients  SourceClientRegistry
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	priorityTask   chan Task
	normalTask     chan Task
	taskProcessors map[string]TaskProcessor
	httpService    http.IService
}

// TaskProcessor is a function that processes a specific task
type TaskProcessor func(task any, sourceClient source.WebSource, spider spider.TaskSpider) (any, error)

// NewProcessor creates a new task processor
func NewProcessor(service *Service, cfg *config.Config, spider *spider.HeadSpider, httpService http.IService) *Processor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Processor{
		service:        service,
		config:         cfg,
		spider:         spider,
		sourceClients:  make(SourceClientRegistry),
		ctx:            ctx,
		cancel:         cancel,
		priorityTask:   make(chan Task, 10),
		normalTask:     make(chan Task, 100),
		taskProcessors: make(map[string]TaskProcessor),
		httpService:    httpService,
	}
}

// RegisterSourceClient registers a source client for a specific source type
func (p *Processor) RegisterSourceClient(sourceType SourceType, client source.WebSource) {
	p.sourceClients[sourceType] = client
}

// RegisterTaskProcessor registers a task processor for a specific task type
func (p *Processor) RegisterTaskProcessor(taskType TaskType, processor TaskProcessor) {
	p.taskProcessors[string(taskType)] = processor
}

// Start starts the task processor
func (p *Processor) Start() {
	p.wg.Add(1)
	go p.processTasksFromQueue()
}

// processTasksFromQueue processes tasks from the queue
func (p *Processor) processTasksFromQueue() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.service.GetTasks():
			if task.Topic == p.config.RabbitMQ.PriorityTopic {
				select {
				case p.priorityTask <- task:
				default:
					p.processTask(task)
				}
			} else {
				select {
				case p.normalTask <- task:
				default:
					p.processTask(task)
				}
			}
		}
	}
}

func (p *Processor) checkAgentActive() (bool, error) {
	return p.httpService.GetAgentService().IsActive(context.Background(), p.httpService.GetAgent().ID.String())
}

// processTask processes a single task
func (p *Processor) processTask(task Task) {
	isActive, err := p.checkAgentActive()
	if err != nil || !isActive {
		logger.Warn().Msg("Agent is not active, skipping task")
		return
	}

	source, taskType, err := ParseTopicInfo(task.Topic)
	if err != nil {
		logger.Error().Err(err).Str("topic", task.Topic).Msg("Error parsing topic")
		return
	}

	// Get the source client
	sourceClient, ok := p.sourceClients[source]
	if !ok {
		logger.Error().Str("source", string(source)).Msg("No source client registered for source")
		return
	}

	// Parse the task
	parsedTask, err := ParseTask(task)
	if err != nil {
		logger.Error().Err(err).Str("topic", task.Topic).Msg("Error parsing task")
		return
	}

	// Get the task processor
	processor, ok := p.taskProcessors[string(taskType)]
	if !ok {
		logger.Error().Str("taskType", string(taskType)).Msg("No task processor registered for task type")
		return
	}

	// Extract task info for reporting
	var url string
	var httpTaskType http.TaskType
	var httpSourceType http.SourceType

	// Convert task type and source to HTTP types
	switch taskType {
	case TaskTypeBook:
		httpTaskType = http.TaskTypeBook
	case TaskTypeChapter:
		httpTaskType = http.TaskTypeChapter
	case TaskTypeSession:
		httpTaskType = http.TaskTypeSession
	}

	switch source {
	case SourceTypeSangTacViet:
		httpSourceType = http.SourceTypeSangTacViet
	case SourceTypeWikiDich:
		httpSourceType = http.SourceTypeWikiDich
	case SourceTypeMetruyenchu:
		httpSourceType = http.SourceTypeMetruyenchu
	}

	// Get URL from task
	switch v := parsedTask.(type) {
	case BookTask:
		url = v.BookURL
	case ChapterTask:
		url = v.ChapterURL
	case SessionTask:
		url = v.URL
	}

	// Generate task ID for reporting
	taskID := ""
	if p.httpService != nil && p.httpService.IsReportingEnabled() {
		taskID = http.GenerateTaskID(httpTaskType, httpSourceType, url)
	}

	// Process the task
	data, err := processor(parsedTask, sourceClient, p.spider)

	// Report task result if control API is configured
	if p.httpService != nil && p.httpService.IsReportingEnabled() {
		taskSvc := p.httpService.GetTaskService()
		ctx := context.Background()

		if err != nil {
			// Report error
			if reportErr := taskSvc.ReportTaskError(ctx, taskID, httpTaskType, httpSourceType, url, err); reportErr != nil {
				logger.Error().Err(reportErr).Str("taskID", taskID).Msg("Error reporting task error")
			}
			logger.Error().Err(err).Str("taskID", taskID).Str("url", url).Msg("Error processing task")
			return
		}

		// Report success
		if reportErr := taskSvc.ReportTaskSuccess(ctx, taskID, httpTaskType, httpSourceType, url, data.(json.RawMessage)); reportErr != nil {
			logger.Error().Err(reportErr).Str("taskID", taskID).Msg("Error reporting task success")
		}
	} else if err != nil {
		logger.Error().Err(err).Str("url", url).Msg("Error processing task")
		return
	}
}

// ProcessTasks processes tasks with priority
func (p *Processor) ProcessTasks() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		for {
			select {
			case <-p.ctx.Done():
				return
			case task := <-p.priorityTask:
				p.processTask(task)
				time.Sleep(3 * time.Second)
			case <-time.After(100 * time.Millisecond):
				select {
				case task := <-p.normalTask:
					p.processTask(task)
					time.Sleep(3 * time.Second)
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

// Stop stops the task processor
func (p *Processor) Stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *Processor) RegisterDefaultTaskProcessors() {
	p.RegisterTaskProcessor(TaskTypeBook, func(task any, sourceClient source.WebSource, spider spider.TaskSpider) (any, error) {
		bookTask, ok := task.(BookTask)
		if !ok {
			return nil, fmt.Errorf("invalid task type, expected BookTask")
		}

		logger.Info().Interface("task", bookTask).Msg("Processing book task")

		// Process the book URL using the spider
		data, err := spider.ProcessPageWithCallback(bookTask.BookURL, sourceClient.ExtractBookInfo)
		if err != nil {
			return nil, fmt.Errorf("error processing book task: %w", err)
		}
		logger.Info().Interface("task", bookTask).Msg("Successfully processed book task")
		return data, nil
	})

	// Register chapter task processor
	p.RegisterTaskProcessor(TaskTypeChapter, func(task interface{}, sourceClient source.WebSource, spider spider.TaskSpider) (any, error) {
		chapterTask, ok := task.(ChapterTask)
		if !ok {
			return nil, fmt.Errorf("invalid task type, expected ChapterTask")
		}

		logger.Info().Interface("task", chapterTask).Msg("Processing chapter task")

		// Process the chapter URL using the spider
		data, err := spider.ProcessPageWithCallback(chapterTask.ChapterURL, sourceClient.ExtractChapter)
		if err != nil {
			return nil, fmt.Errorf("error processing chapter task: %w", err)
		}

		logger.Info().Interface("task", chapterTask).Msg("Successfully processed chapter task")
		return data, nil
	})

	// Register session task processor
	p.RegisterTaskProcessor(TaskTypeSession, func(task interface{}, sourceClient source.WebSource, spider spider.TaskSpider) (any, error) {
		sessionTask, ok := task.(SessionTask)
		if !ok {
			return nil, fmt.Errorf("invalid task type, expected SessionTask")
		}

		logger.Info().Interface("task", sessionTask).Msg("Processing session task")
		// Process the session URL using the spider
		data, err := spider.ProcessPageWithCallback(sessionTask.URL, sourceClient.ExtractSession)
		if err != nil {
			return nil, fmt.Errorf("error processing session task: %w", err)
		}

		logger.Info().Interface("task", sessionTask).Msg("Successfully processed session task")
		return data, nil
	})
}
