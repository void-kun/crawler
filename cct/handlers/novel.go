package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cct/models"
)

// GetNovels handles GET /novels
func GetNovels(w http.ResponseWriter, r *http.Request) {
	// Check if website_id query parameter is provided
	websiteIDStr := r.URL.Query().Get("website_id")

	var novels []models.Novel
	var err error

	if websiteIDStr != "" {
		// Get novels for a specific website
		websiteID, err := strconv.Atoi(websiteIDStr)
		if err != nil {
			http.Error(w, "Invalid website ID", http.StatusBadRequest)
			return
		}

		novels, err = models.GetNovelsByWebsite(websiteID)
	} else {
		// Get all novels
		novels, err = models.GetNovels()
	}

	if err != nil {
		http.Error(w, "Failed to get novels: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(novels)
}

// GetNovel handles GET /novels/{id}
func GetNovel(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid novel ID", http.StatusBadRequest)
		return
	}

	novel, err := models.GetNovel(id)
	if err != nil {
		http.Error(w, "Failed to get novel: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(novel)
}

// CreateNovel handles POST /novels
func CreateNovel(w http.ResponseWriter, r *http.Request) {
	var novel models.Novel
	if err := json.NewDecoder(r.Body).Decode(&novel); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := models.CreateNovel(&novel); err != nil {
		http.Error(w, "Failed to create novel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(novel)
}

// UpdateNovel handles PUT /novels/{id}
func UpdateNovel(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid novel ID", http.StatusBadRequest)
		return
	}

	var novel models.Novel
	if err := json.NewDecoder(r.Body).Decode(&novel); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	novel.ID = id

	if err := models.UpdateNovel(&novel); err != nil {
		http.Error(w, "Failed to update novel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(novel)
}

// DeleteNovel handles DELETE /novels/{id}
func DeleteNovel(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid novel ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteNovel(id); err != nil {
		http.Error(w, "Failed to delete novel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
