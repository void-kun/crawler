package main

import (
	"log" // Standard log for initialization only
	"net/http"
	"strconv"
	"time"

	"cct/config"
	"cct/handlers"
	"cct/middleware"
	"cct/pkg/logger"
	"cct/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Use standard log for initialization errors
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	if err := logger.Init(&cfg.Logging); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize database connection
	if err := utils.InitDB(&cfg.Database); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer utils.CloseDB()

	// Initialize RabbitMQ service
	if err := handlers.InitRabbitMQService(cfg); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize RabbitMQ service")
	}
	defer handlers.CloseRabbitMQService()

	// Create a new router
	mux := http.NewServeMux()

	// API routes
	// Websites
	mux.HandleFunc("GET /api/websites", handlers.GetWebsites)
	mux.HandleFunc("GET /api/websites/{id}", handlers.GetWebsite)
	mux.HandleFunc("POST /api/websites", handlers.CreateWebsite)
	mux.HandleFunc("PUT /api/websites/{id}", handlers.UpdateWebsite)
	mux.HandleFunc("DELETE /api/websites/{id}", handlers.DeleteWebsite)

	// Novels
	mux.HandleFunc("GET /api/novels", handlers.GetNovels)
	mux.HandleFunc("GET /api/novels/{id}", handlers.GetNovel)
	mux.HandleFunc("POST /api/novels", handlers.CreateNovel)
	mux.HandleFunc("PUT /api/novels/{id}", handlers.UpdateNovel)
	mux.HandleFunc("DELETE /api/novels/{id}", handlers.DeleteNovel)

	// Chapters
	mux.HandleFunc("GET /api/chapters", handlers.GetChapters)
	mux.HandleFunc("GET /api/chapters/{id}", handlers.GetChapter)
	mux.HandleFunc("POST /api/chapters", handlers.CreateChapter)
	mux.HandleFunc("PUT /api/chapters/{id}", handlers.UpdateChapter)
	mux.HandleFunc("DELETE /api/chapters/{id}", handlers.DeleteChapter)

	// Crawl Jobs
	mux.HandleFunc("GET /api/crawl-jobs", handlers.GetCrawlJobs)
	mux.HandleFunc("GET /api/crawl-jobs/{id}", handlers.GetCrawlJob)
	mux.HandleFunc("POST /api/crawl-jobs", handlers.CreateCrawlJob)
	mux.HandleFunc("PUT /api/crawl-jobs/{id}", handlers.UpdateCrawlJob)
	// mux.HandleFunc("DELETE /api/crawl-jobs/{id}", handlers.DeleteCrawlJob)
	// mux.HandleFunc("POST /api/crawl-jobs/{id}/start", handlers.StartCrawlJob)
	// mux.HandleFunc("POST /api/crawl-jobs/{id}/complete", handlers.CompleteCrawlJob)
	// mux.HandleFunc("POST /api/crawl-jobs/{id}/fail", handlers.FailCrawlJob)

	// Agents
	mux.HandleFunc("GET /api/agents", handlers.GetAgents)
	mux.HandleFunc("GET /api/agents/{id}", handlers.GetAgent)
	mux.HandleFunc("POST /api/agents", handlers.CreateAgent)
	mux.HandleFunc("PUT /api/agents/{id}", handlers.UpdateAgent)
	mux.HandleFunc("DELETE /api/agents/{id}", handlers.DeleteAgent)
	mux.HandleFunc("POST /api/agents/{id}/heartbeat", handlers.HeartbeatAgent)
	mux.HandleFunc("POST /api/agents/deactivate-inactive", handlers.DeactivateInactiveAgents)

	// Users
	mux.HandleFunc("GET /api/users", handlers.GetUsers)
	mux.HandleFunc("GET /api/users/{id}", handlers.GetUser)
	mux.HandleFunc("POST /api/users", handlers.CreateUser)
	mux.HandleFunc("PUT /api/users/{id}", handlers.UpdateUser)
	mux.HandleFunc("PUT /api/users/{id}/password", handlers.UpdateUserPassword)
	mux.HandleFunc("DELETE /api/users/{id}", handlers.DeleteUser)

	// RabbitMQ Tasks
	mux.HandleFunc("POST /api/tasks/publish", handlers.PublishTask)
	max.HandleFunc("POST /api/tasks/result", handlers.ResultTask)

	// Authentication
	mux.HandleFunc("POST /api/auth/login", handlers.Login)
	mux.HandleFunc("POST /api/auth/register", handlers.Register)

	// Apply middleware
	var handler http.Handler = mux

	// Enable authentication if configured
	if cfg.Auth.Enabled {
		handler = middleware.AuthMiddleware(handler)
	}

	// Add logging middleware
	handler = loggingMiddleware(handler)

	// Get port from configuration
	port := strconv.Itoa(cfg.Server.Port)

	// Create server with timeouts from configuration
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server
	logger.Info().Str("port", port).Msg("Server starting")
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(lw, r)

		// Log the request
		duration := time.Since(start)
		logger.Info().
			Str("method", r.Method).
			Str("path", r.RequestURI).
			Int("status", lw.statusCode).
			Dur("duration ms", duration).
			Msg("Request processed")
	})
}

// loggingResponseWriter is a custom response writer that captures the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}
