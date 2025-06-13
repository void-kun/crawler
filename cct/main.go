package main

import (
	"log" // Standard log for initialization only
	"net/http"
	"strconv"
	"strings"
	"time"

	"cct/config"
	"cct/handlers"
	"cct/middleware"
	"cct/pkg/logger"
	"cct/pkg/scheduler"
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

	// Initialize and start scheduler if enabled
	if cfg.Scheduler.Enabled {
		agentService, err := handlers.GetAgentService()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to get agent service for scheduler")
		}

		schedulerService := scheduler.NewScheduler(cfg, agentService)
		if err := schedulerService.Start(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start scheduler")
		}
		defer schedulerService.Stop()

		logger.Info().Msg("Scheduler started successfully")
	}

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

	// Schedules
	mux.HandleFunc("GET /api/schedules", handlers.GetSchedules)
	mux.HandleFunc("GET /api/schedules/due", handlers.GetDueSchedules)
	mux.HandleFunc("GET /api/schedules/{id}", handlers.GetSchedule)
	mux.HandleFunc("POST /api/schedules", handlers.CreateSchedule)
	mux.HandleFunc("PUT /api/schedules/{id}", handlers.UpdateSchedule)
	mux.HandleFunc("DELETE /api/schedules/{id}", handlers.DeleteSchedule)
	mux.HandleFunc("POST /api/schedules/{id}/trigger", handlers.TriggerSchedule)

	// Chapter crawl logs
	mux.HandleFunc("GET /api/chapters/{id}/logs", handlers.GetChapterCrawlLogs)

	// RabbitMQ Tasks
	mux.HandleFunc("POST /api/tasks/publish", handlers.PublishTask)
	mux.HandleFunc("POST /api/tasks/result", handlers.ResultTask)

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

var ignoreLogPaths = []string{
	"heartbeat",
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, path := range ignoreLogPaths {
			if strings.HasSuffix(r.RequestURI, path) {
				next.ServeHTTP(w, r)
				return
			}
		}
		start := time.Now()
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lw, r)
		logger.Info().
			Str("method", r.Method).
			Str("path", r.RequestURI).
			Int("status", lw.statusCode).
			Dur("duration ms", time.Since(start)).
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
