package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cct/config"
	"cct/models"
	"cct/pkg/logger"
	"cct/pkg/rabbitmq"
)

var agentService *rabbitmq.AgentService

type Book struct {
	BookUrl      string
	BookId       string
	BookName     string
	BookImageUrl string
	AuthorName   string
	Chapters     []Chapter
	BookHost     string
}

type Chapter struct {
	ChapterId     string
	ChapterName   string
	ChapterUrl    string
	ChapterNumber int
}

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

// GetAgentService returns the agent service instance
func GetAgentService() (*rabbitmq.AgentService, error) {
	if agentService == nil {
		return nil, fmt.Errorf("agent service not initialized")
	}
	return agentService, nil
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
	Data        json.RawMessage     `json:"data,omitempty"`
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
	if agentService == nil {
		http.Error(w, "RabbitMQ service not initialized", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req TaskResult
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.TaskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	if req.TaskType == "" {
		http.Error(w, "Task type is required", http.StatusBadRequest)
		return
	}

	websites, err := models.GetWebsites()
	if len(websites) == 0 || err != nil {
		http.Error(w, "No websites found", http.StatusInternalServerError)
		return
	}

	// Get website by SourceType
	var website models.Website
	for _, w := range websites {
		if rabbitmq.SourceType(w.Name) == req.Source {
			website = w
			break
		}
	}

	logger.Debug().Interface("result", req).Msg("Received task result")
	// Now you can use the data variable which holds the unmarshaled Book or Chapter struct
	switch req.TaskType {
	case rabbitmq.TaskTypeBook:
		var book Book
		if err := json.Unmarshal(req.Data, &book); err != nil {
			http.Error(w, "Invalid data: "+err.Error(), http.StatusBadRequest)
			return
		}
		logger.Info().Interface("book with chapters", len(book.Chapters)).Msg("Book data received")

		// Get novel
		existedNovel, err := models.GetNovelByUrl(book.BookUrl)
		if err != nil {
			logger.Error().Err(err).Str("url", book.BookUrl).Msg("Failed to get novel")
		}
		logger.Info().Interface("novel", existedNovel).Msg("Novel data received")
		if existedNovel.ID != 0 {
			logger.Info().Interface("novel", existedNovel).Msg("Novel already exists")
			logger.Info().Msg("Updating novel")

			// Update novel
			// =========================================================================================================================
			novel := mappingUpdateBookToNovel(book, existedNovel.ID)
			if err := models.UpdateNovel(novel); err != nil {
				http.Error(w, "Failed to update novel: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Update chapters
			var updatedChapters []models.Chapter
			for _, c := range book.Chapters {
				chapter := mappingUpdateChapterToChapter(c, existedNovel.ID)
				if err := models.InsertOrUpdateChapter(chapter); err != nil {
					http.Error(w, "Failed to insert or update chapter: "+err.Error(), http.StatusInternalServerError)
					return
				}
				// Convert to models.Chapter for scheduler processing
				modelChapter := models.Chapter{
					ID:            chapter.ID,
					NovelID:       chapter.NovelID,
					ExternalID:    c.ChapterId,
					Title:         c.ChapterName,
					URL:           c.ChapterUrl,
					ChapterNumber: c.ChapterNumber,
					Content:       chapter.Content,
					CrawledAt:     chapter.CrawledAt,
					Error:         chapter.Error,
				}
				updatedChapters = append(updatedChapters, modelChapter)
			}

			// Process book crawl result for scheduler (create chapter crawl jobs)
			processBookCrawlForScheduler(existedNovel.ID, updatedChapters)
			return
		} else {
			logger.Info().Msg("Creating novel")
			// Create novel
			// =========================================================================================================================
			novel := models.Novel{
				WebsiteID:  website.ID,
				ExternalID: book.BookId,
				Title:      book.BookName,
				SourceURL:  book.BookUrl,
				LastCrawledAt: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
			}
			if err := models.CreateNovel(&novel); err != nil {
				http.Error(w, "Failed to create novel: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Create chapters
			var createdChapters []models.Chapter
			for _, c := range book.Chapters {
				chapter := models.Chapter{
					NovelID:       novel.ID,
					ExternalID:    c.ChapterId,
					Title:         c.ChapterName,
					URL:           c.ChapterUrl,
					ChapterNumber: c.ChapterNumber,
					CrawledAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}
				if err := models.CreateChapter(&chapter); err != nil {
					http.Error(w, "Failed to create chapter: "+err.Error(), http.StatusInternalServerError)
					return
				}
				createdChapters = append(createdChapters, chapter)
			}

			// Process book crawl result for scheduler (create chapter crawl jobs)
			processBookCrawlForScheduler(novel.ID, createdChapters)
		}

	case rabbitmq.TaskTypeChapter:
		var chapterContent string
		if err := json.Unmarshal(req.Data, &chapterContent); err != nil {
			http.Error(w, "Invalid data: "+err.Error(), http.StatusBadRequest)
			return
		}
		logger.Info().Interface("chapter", req.URL).Msg("Chapter data received")

		updateErr := models.UpdateChapterByUrl(req.URL, chapterContent)
		if updateErr != nil {
			// Log chapter crawl failure
			if chapter, err := models.GetChapterByUrl(req.URL); err == nil {
				logChapterCrawlResult(chapter.ID, false, updateErr.Error())
			}
			http.Error(w, "Failed to update chapter content: "+updateErr.Error(), http.StatusInternalServerError)
			return
		}

		// Log successful chapter crawl
		if chapter, err := models.GetChapterByUrl(req.URL); err == nil {
			logChapterCrawlResult(chapter.ID, true, "")
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Task result received successfully",
	})
}

func mappingUpdateBookToNovel(book Book, novelID int) *models.Novel {
	novel := &models.Novel{
		ID: novelID,
		LastCrawledAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	if book.BookName != "" {
		novel.Title = book.BookName
	}
	if book.BookUrl != "" {
		novel.SourceURL = book.BookUrl
	}

	return novel
}

func mappingUpdateChapterToChapter(c Chapter, novelID int) *models.Chapter {
	chapter := &models.Chapter{
		NovelID: novelID,
		CrawledAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	if c.ChapterName != "" {
		chapter.Title = c.ChapterName
	}
	if c.ChapterUrl != "" {
		chapter.URL = c.ChapterUrl
	}
	return chapter
}

// processBookCrawlForScheduler processes book crawl results for scheduler
func processBookCrawlForScheduler(novelID int, chapters []models.Chapter) {
	// Update novel's last_crawled_at
	err := models.UpdateNovelLastCrawledAt(novelID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("novel_id", novelID).
			Msg("Failed to update novel last_crawled_at")
		return
	}

	// Get the novel to determine the website
	novel, err := models.GetNovel(novelID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("novel_id", novelID).
			Msg("Failed to get novel for chapter crawling")
		return
	}

	// Get the website to determine the source type
	website, err := models.GetWebsite(novel.WebsiteID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("website_id", novel.WebsiteID).
			Msg("Failed to get website for chapter crawling")
		return
	}

	sourceType := rabbitmq.SourceType(website.Name)

	// Create chapter crawl tasks for chapters that need content
	for _, chapter := range chapters {
		// Check if chapter needs crawling (no content or failed status)
		if chapter.Content == "" || chapter.Error != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			err = agentService.PublishChapterTask(ctx, sourceType, chapter.URL)
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
		Msg("Processed book crawl result for scheduler")
}

// logChapterCrawlResult logs the result of a chapter crawl
func logChapterCrawlResult(chapterID int, success bool, errorMsg string) {
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
		return
	}

	logger.Info().
		Int("chapter_id", chapterID).
		Str("status", status).
		Msg("Logged chapter crawl result")
}
