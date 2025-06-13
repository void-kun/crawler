package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cct/models"
)

// GetSchedules handles GET /schedules
func GetSchedules(w http.ResponseWriter, r *http.Request) {
	// Check if novel_id query parameter is provided
	novelIDStr := r.URL.Query().Get("novel_id")

	var schedules []models.NovelSchedule
	var err error

	if novelIDStr != "" {
		// Get schedules for a specific novel
		var novelID int
		novelID, err = strconv.Atoi(novelIDStr)
		if err != nil {
			http.Error(w, "Invalid novel ID", http.StatusBadRequest)
			return
		}

		schedules, err = models.GetSchedulesByNovel(novelID)
	} else {
		// Get all schedules
		schedules, err = models.GetSchedules()
	}

	if err != nil {
		http.Error(w, "Failed to get schedules: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

// GetSchedule handles GET /schedules/{id}
func GetSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	schedule, err := models.GetSchedule(id)
	if err != nil {
		http.Error(w, "Failed to get schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// CreateScheduleRequest represents the request body for creating a schedule
type CreateScheduleRequest struct {
	NovelID         int  `json:"novel_id"`
	Enabled         bool `json:"enabled"`
	IntervalSeconds int  `json:"interval_seconds"`
}

// CreateSchedule handles POST /schedules
func CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.NovelID <= 0 {
		http.Error(w, "Novel ID is required and must be positive", http.StatusBadRequest)
		return
	}
	if req.IntervalSeconds <= 0 {
		http.Error(w, "Interval seconds is required and must be positive", http.StatusBadRequest)
		return
	}

	// Check if novel exists
	_, err := models.GetNovel(req.NovelID)
	if err != nil {
		http.Error(w, "Novel not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create schedule
	schedule := &models.NovelSchedule{
		NovelID:         req.NovelID,
		Enabled:         req.Enabled,
		IntervalSeconds: req.IntervalSeconds,
	}

	if err := models.CreateSchedule(schedule); err != nil {
		http.Error(w, "Failed to create schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

// UpdateScheduleRequest represents the request body for updating a schedule
type UpdateScheduleRequest struct {
	Enabled         *bool `json:"enabled,omitempty"`
	IntervalSeconds *int  `json:"interval_seconds,omitempty"`
}

// UpdateSchedule handles PUT /schedules/{id}
func UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	var req UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get existing schedule
	schedule, err := models.GetSchedule(id)
	if err != nil {
		http.Error(w, "Failed to get schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.Enabled != nil {
		schedule.Enabled = *req.Enabled
	}
	if req.IntervalSeconds != nil {
		if *req.IntervalSeconds <= 0 {
			http.Error(w, "Interval seconds must be positive", http.StatusBadRequest)
			return
		}
		schedule.IntervalSeconds = *req.IntervalSeconds
	}

	if err := models.UpdateSchedule(&schedule); err != nil {
		http.Error(w, "Failed to update schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// DeleteSchedule handles DELETE /schedules/{id}
func DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteSchedule(id); err != nil {
		http.Error(w, "Failed to delete schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TriggerSchedule handles POST /schedules/{id}/trigger
func TriggerSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	// Check if schedule exists
	_, err = models.GetSchedule(id)
	if err != nil {
		http.Error(w, "Failed to get schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update next_run_at to current time to trigger immediate execution
	err = models.TriggerScheduleNow(id)
	if err != nil {
		http.Error(w, "Failed to trigger schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated schedule
	updatedSchedule, err := models.GetSchedule(id)
	if err != nil {
		http.Error(w, "Failed to get updated schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "success",
		"message":  "Schedule triggered successfully",
		"schedule": updatedSchedule,
	})
}

// GetDueSchedules handles GET /schedules/due
func GetDueSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := models.GetDueSchedules()
	if err != nil {
		http.Error(w, "Failed to get due schedules: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"count":     len(schedules),
		"schedules": schedules,
	})
}

// GetChapterCrawlLogs handles GET /chapters/{id}/logs
func GetChapterCrawlLogs(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	chapterID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}

	logs, err := models.GetChapterCrawlLogs(chapterID)
	if err != nil {
		http.Error(w, "Failed to get chapter crawl logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
