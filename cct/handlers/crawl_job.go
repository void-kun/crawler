package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cct/models"
)

// GetCrawlJobs handles GET /crawl-jobs
func GetCrawlJobs(w http.ResponseWriter, r *http.Request) {
	// Check if novel_id query parameter is provided
	novelIDStr := r.URL.Query().Get("novel_id")

	var jobs []models.CrawlJob
	var err error

	if novelIDStr != "" {
		// Get crawl jobs for a specific novel
		novelID, err := strconv.Atoi(novelIDStr)
		if err != nil {
			http.Error(w, "Invalid novel ID", http.StatusBadRequest)
			return
		}

		jobs, err = models.GetCrawlJobsByNovel(novelID)
	} else {
		// Get all crawl jobs
		jobs, err = models.GetCrawlJobs()
	}

	if err != nil {
		http.Error(w, "Failed to get crawl jobs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// GetCrawlJob handles GET /crawl-jobs/{id}
func GetCrawlJob(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crawl job ID", http.StatusBadRequest)
		return
	}

	job, err := models.GetCrawlJob(id)
	if err != nil {
		http.Error(w, "Failed to get crawl job: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// CreateCrawlJob handles POST /crawl-jobs
func CreateCrawlJob(w http.ResponseWriter, r *http.Request) {
	var job models.CrawlJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set default status if not provided
	if job.Status == "" {
		job.Status = "pending"
	}

	if err := models.CreateCrawlJob(&job); err != nil {
		http.Error(w, "Failed to create crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// UpdateCrawlJob handles PUT /crawl-jobs/{id}
func UpdateCrawlJob(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crawl job ID", http.StatusBadRequest)
		return
	}

	var job models.CrawlJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	job.ID = id

	if err := models.UpdateCrawlJob(&job); err != nil {
		http.Error(w, "Failed to update crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// StartCrawlJob handles POST /crawl-jobs/{id}/start
func StartCrawlJob(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crawl job ID", http.StatusBadRequest)
		return
	}

	if err := models.StartCrawlJob(id); err != nil {
		http.Error(w, "Failed to start crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	job, err := models.GetCrawlJob(id)
	if err != nil {
		http.Error(w, "Failed to get updated crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// CompleteCrawlJob handles POST /crawl-jobs/{id}/complete
func CompleteCrawlJob(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crawl job ID", http.StatusBadRequest)
		return
	}

	if err := models.CompleteCrawlJob(id); err != nil {
		http.Error(w, "Failed to complete crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	job, err := models.GetCrawlJob(id)
	if err != nil {
		http.Error(w, "Failed to get updated crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// FailCrawlJob handles POST /crawl-jobs/{id}/fail
func FailCrawlJob(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crawl job ID", http.StatusBadRequest)
		return
	}

	// Parse error message from request body
	var requestBody struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := models.FailCrawlJob(id, requestBody.Error); err != nil {
		http.Error(w, "Failed to fail crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	job, err := models.GetCrawlJob(id)
	if err != nil {
		http.Error(w, "Failed to get updated crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// DeleteCrawlJob handles DELETE /crawl-jobs/{id}
func DeleteCrawlJob(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crawl job ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteCrawlJob(id); err != nil {
		http.Error(w, "Failed to delete crawl job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
