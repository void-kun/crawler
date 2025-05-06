package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zrik/agent/appagent/internal/source"
	"github.com/zrik/agent/appagent/pkg/config"
	http "github.com/zrik/agent/appagent/pkg/http"
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
	httpService    *http.Service
}

// TaskProcessor is a function that processes a specific task
type TaskProcessor func(task any, sourceClient source.WebSource, spider spider.TaskSpider) (any, error)

// NewProcessor creates a new task processor
func NewProcessor(service *Service, cfg *config.Config, spider *spider.HeadSpider) *Processor {
	ctx, cancel := context.WithCancel(context.Background())

	// Create HTTP service if control API is configured
	var httpService *http.Service
	if cfg.ControlAPI.BaseURL != "" {
		httpService = http.NewService(&cfg.ControlAPI)
	}

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

// processTask processes a single task
func (p *Processor) processTask(task Task) {
	source, taskType, err := ParseTopicInfo(task.Topic)
	if err != nil {
		log.Printf("Error parsing topic: %v", err)
		return
	}

	// Get the source client
	sourceClient, ok := p.sourceClients[source]
	if !ok {
		log.Printf("No source client registered for source: %s", source)
		return
	}

	// Parse the task
	parsedTask, err := ParseTask(task)
	if err != nil {
		log.Printf("Error parsing task: %v", err)
		return
	}

	// Get the task processor
	processor, ok := p.taskProcessors[string(taskType)]
	if !ok {
		log.Printf("No task processor registered for task type: %s", taskType)
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
		taskResultSvc := p.httpService.GetTaskResultService()
		ctx := context.Background()

		if err != nil {
			// Report error
			if reportErr := taskResultSvc.ReportTaskError(ctx, taskID, httpTaskType, httpSourceType, url, err); reportErr != nil {
				log.Printf("Error reporting task error: %v", reportErr)
			}
			log.Printf("Error processing task: %v", err)
			return
		}

		// Report success
		if reportErr := taskResultSvc.ReportTaskSuccess(ctx, taskID, httpTaskType, httpSourceType, url, data); reportErr != nil {
			log.Printf("Error reporting task success: %v", reportErr)
		}
	} else if err != nil {
		log.Printf("Error processing task: %v", err)
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
			case <-time.After(100 * time.Millisecond):
				select {
				case task := <-p.normalTask:
					p.processTask(task)
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

		log.Printf("Processing book task: %+v", bookTask)

		// Process the book URL using the spider
		data, err := spider.ProcessPageWithCallback(bookTask.BookURL, sourceClient.ExtractBookInfo)
		if err != nil {
			return nil, fmt.Errorf("error processing book task: %w", err)
		}

		log.Printf("Successfully processed book task: %+v", bookTask)
		return data, nil
	})

	// Register chapter task processor
	p.RegisterTaskProcessor(TaskTypeChapter, func(task interface{}, sourceClient source.WebSource, spider spider.TaskSpider) (any, error) {
		chapterTask, ok := task.(ChapterTask)
		if !ok {
			return nil, fmt.Errorf("invalid task type, expected ChapterTask")
		}

		log.Printf("Processing chapter task: %+v", chapterTask)

		// Process the chapter URL using the spider
		data, err := spider.ProcessPageWithCallback(chapterTask.ChapterURL, sourceClient.ExtractChapter)
		if err != nil {
			return nil, fmt.Errorf("error processing chapter task: %w", err)
		}

		log.Printf("Successfully processed chapter task: %+v", chapterTask)
		return data, nil
	})

	// Register session task processor
	p.RegisterTaskProcessor(TaskTypeSession, func(task interface{}, sourceClient source.WebSource, spider spider.TaskSpider) (any, error) {
		sessionTask, ok := task.(SessionTask)
		if !ok {
			return nil, fmt.Errorf("invalid task type, expected SessionTask")
		}

		log.Printf("Processing session task: %+v", sessionTask)

		p.spider.SetHeadless(false)
		// Process the session URL using the spider
		data, err := spider.ProcessPageWithCallback(sessionTask.URL, sourceClient.ExtractSession)
		if err != nil {
			return nil, fmt.Errorf("error processing session task: %w", err)
		}

		p.spider.SetHeadless(true)
		log.Printf("Successfully processed session task: %+v", sessionTask)
		return data, nil
	})
}
