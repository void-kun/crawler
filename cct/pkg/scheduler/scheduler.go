package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cct/config"
	"cct/models"
	"cct/pkg/logger"
	"cct/pkg/rabbitmq"
)

// Scheduler manages scheduled crawl tasks
type Scheduler struct {
	agentService *rabbitmq.AgentService
	config       *config.Config
	ticker       *time.Ticker
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler(cfg *config.Config, agentService *rabbitmq.AgentService) *Scheduler {
	return &Scheduler{
		agentService: agentService,
		config:       cfg,
		stopCh:       make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	// Default to 1 minute interval if not configured
	interval := 1 * time.Minute
	if s.config.Scheduler.CheckInterval > 0 {
		interval = time.Duration(s.config.Scheduler.CheckInterval) * time.Second
	}

	s.ticker = time.NewTicker(interval)
	s.running = true

	s.wg.Add(1)
	go s.run()

	logger.Info().Dur("interval", interval).Msg("Scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	close(s.stopCh)
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.running = false

	s.wg.Wait()
	logger.Info().Msg("Scheduler stopped")
	return nil
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ticker.C:
			logger.Info().Msg("Scheduler: checking for due schedules")
			s.processDueSchedules()
		case <-s.stopCh:
			return
		}
	}
}

// processDueSchedules processes all schedules that are due to run
func (s *Scheduler) processDueSchedules() {
	schedules, err := models.GetDueSchedules()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get due schedules")
		return
	}

	if len(schedules) == 0 {
		logger.Debug().Msg("No due schedules found")
		return
	}

	logger.Info().Int("count", len(schedules)).Msg("Processing due schedules")

	for _, schedule := range schedules {
		s.processSchedule(schedule)
	}
}

// processSchedule processes a single schedule
func (s *Scheduler) processSchedule(schedule models.NovelSchedule) {
	logger.Info().
		Int("schedule_id", schedule.ID).
		Int("novel_id", schedule.NovelID).
		Msg("Processing schedule")

	// Get the novel to crawl
	novel, err := models.GetNovel(schedule.NovelID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("schedule_id", schedule.ID).
			Int("novel_id", schedule.NovelID).
			Msg("Failed to get novel for schedule")
		return
	}

	// Get the website to determine the source type
	website, err := models.GetWebsite(novel.WebsiteID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("schedule_id", schedule.ID).
			Int("website_id", novel.WebsiteID).
			Msg("Failed to get website for schedule")
		return
	}

	// Convert website name to source type
	sourceType := rabbitmq.SourceType(website.Name)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Publish book crawl task
	err = s.agentService.PublishBookTask(ctx, sourceType, novel.SourceURL)
	if err != nil {
		logger.Error().
			Err(err).
			Int("schedule_id", schedule.ID).
			Int("novel_id", schedule.NovelID).
			Str("source_url", novel.SourceURL).
			Msg("Failed to publish book task")
		return
	}

	// Update schedule run times
	err = models.UpdateScheduleRunTime(schedule.ID, schedule.IntervalSeconds)
	if err != nil {
		logger.Error().
			Err(err).
			Int("schedule_id", schedule.ID).
			Msg("Failed to update schedule run time")
		return
	}

	logger.Info().
		Int("schedule_id", schedule.ID).
		Int("novel_id", schedule.NovelID).
		Str("source_url", novel.SourceURL).
		Msg("Successfully processed schedule")
}

// ProcessBookCrawlResult processes the result of a book crawl task
func (s *Scheduler) ProcessBookCrawlResult(novelID int, chapters []models.Chapter) error {
	// Update novel's last_crawled_at
	err := models.UpdateNovelLastCrawledAt(novelID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("novel_id", novelID).
			Msg("Failed to update novel last_crawled_at")
		return err
	}

	// Get the novel to determine the website
	novel, err := models.GetNovel(novelID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("novel_id", novelID).
			Msg("Failed to get novel for chapter crawling")
		return err
	}

	// Get the website to determine the source type
	website, err := models.GetWebsite(novel.WebsiteID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("website_id", novel.WebsiteID).
			Msg("Failed to get website for chapter crawling")
		return err
	}

	sourceType := rabbitmq.SourceType(website.Name)

	// Create chapter crawl tasks for chapters that need content
	for _, chapter := range chapters {
		// Check if chapter needs crawling (no content or failed status)
		if chapter.Content == "" || chapter.Error != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			err = s.agentService.PublishChapterTask(ctx, sourceType, chapter.URL)
			if err != nil {
				logger.Error().
					Err(err).
					Int("chapter_id", chapter.ID).
					Str("chapter_url", chapter.URL).
					Msg("Failed to publish chapter task")
				cancel()
				continue
			}

			cancel()

			logger.Info().
				Int("chapter_id", chapter.ID).
				Str("chapter_url", chapter.URL).
				Msg("Published chapter crawl task")
		}
	}

	logger.Info().
		Int("novel_id", novelID).
		Int("total_chapters", len(chapters)).
		Msg("Processed book crawl result")

	return nil
}

// LogChapterCrawlResult logs the result of a chapter crawl
func (s *Scheduler) LogChapterCrawlResult(chapterID int, success bool, errorMsg string) error {
	status := "success"
	if !success {
		status = "failed"
	}

	log := &models.ChapterCrawlLog{
		ChapterID: chapterID,
		Status:    status,
		Error:     errorMsg,
	}

	err := models.CreateChapterCrawlLog(log)
	if err != nil {
		logger.Error().
			Err(err).
			Int("chapter_id", chapterID).
			Str("status", status).
			Msg("Failed to create chapter crawl log")
		return err
	}

	logger.Info().
		Int("chapter_id", chapterID).
		Str("status", status).
		Msg("Logged chapter crawl result")

	return nil
}
