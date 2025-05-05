package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zrik/agent/appagent/internal/source"
	"github.com/zrik/agent/appagent/pkg/config"
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
}

// TaskProcessor is a function that processes a specific task
type TaskProcessor func(task any, sourceClient source.WebSource, spider spider.TaskSpider) error

// NewProcessor creates a new task processor
func NewProcessor(service *Service, cfg *config.Config, spider *spider.HeadSpider) *Processor {
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

	// Process the task
	if err := processor(parsedTask, sourceClient, p.spider); err != nil {
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
	p.RegisterTaskProcessor(TaskTypeBook, func(task any, sourceClient source.WebSource, spider spider.TaskSpider) error {
		bookTask, ok := task.(BookTask)
		if !ok {
			return fmt.Errorf("invalid task type, expected BookTask")
		}

		log.Printf("Processing book task: %+v", bookTask)

		// Process the book URL using the spider
		err := spider.ProcessPageWithCallback(bookTask.BookURL, sourceClient.ExtractBookInfo)
		if err != nil {
			return fmt.Errorf("error processing book task: %w", err)
		}

		log.Printf("Successfully processed book task: %+v", bookTask)
		return nil
	})

	// Register chapter task processor
	p.RegisterTaskProcessor(TaskTypeChapter, func(task interface{}, sourceClient source.WebSource, spider spider.TaskSpider) error {
		chapterTask, ok := task.(ChapterTask)
		if !ok {
			return fmt.Errorf("invalid task type, expected ChapterTask")
		}

		log.Printf("Processing chapter task: %+v", chapterTask)

		// Process the chapter URL using the spider
		err := spider.ProcessPageWithCallback(chapterTask.ChapterURL, sourceClient.ExtractChapter)
		if err != nil {
			return fmt.Errorf("error processing chapter task: %w", err)
		}

		log.Printf("Successfully processed chapter task: %+v", chapterTask)
		return nil
	})

	// Register session task processor
	p.RegisterTaskProcessor(TaskTypeSession, func(task interface{}, sourceClient source.WebSource, spider spider.TaskSpider) error {
		sessionTask, ok := task.(SessionTask)
		if !ok {
			return fmt.Errorf("invalid task type, expected SessionTask")
		}

		log.Printf("Processing session task: %+v", sessionTask)

		p.spider.SetHeadless(false)
		// Process the session URL using the spider
		err := spider.ProcessPageWithCallback(sessionTask.URL, sourceClient.ExtractSession)
		if err != nil {
			return fmt.Errorf("error processing session task: %w", err)
		}

		p.spider.SetHeadless(true)
		log.Printf("Successfully processed session task: %+v", sessionTask)
		return nil
	})
}
