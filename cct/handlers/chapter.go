package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cct/models"
)

// GetChapters handles GET /chapters
func GetChapters(w http.ResponseWriter, r *http.Request) {
	// Check if novel_id query parameter is provided
	novelIDStr := r.URL.Query().Get("novel_id")
	
	var chapters []models.Chapter
	var err error
	
	if novelIDStr != "" {
		// Get chapters for a specific novel
		novelID, err := strconv.Atoi(novelIDStr)
		if err != nil {
			http.Error(w, "Invalid novel ID", http.StatusBadRequest)
			return
		}
		
		chapters, err = models.GetChaptersByNovel(novelID)
	} else {
		// Get all chapters
		chapters, err = models.GetChapters()
	}
	
	if err != nil {
		http.Error(w, "Failed to get chapters: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chapters)
}

// GetChapter handles GET /chapters/{id}
func GetChapter(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}

	chapter, err := models.GetChapter(id)
	if err != nil {
		http.Error(w, "Failed to get chapter: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chapter)
}

// CreateChapter handles POST /chapters
func CreateChapter(w http.ResponseWriter, r *http.Request) {
	var chapter models.Chapter
	if err := json.NewDecoder(r.Body).Decode(&chapter); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := models.CreateChapter(&chapter); err != nil {
		http.Error(w, "Failed to create chapter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chapter)
}

// UpdateChapter handles PUT /chapters/{id}
func UpdateChapter(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}

	var chapter models.Chapter
	if err := json.NewDecoder(r.Body).Decode(&chapter); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	chapter.ID = id

	if err := models.UpdateChapter(&chapter); err != nil {
		http.Error(w, "Failed to update chapter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chapter)
}

// DeleteChapter handles DELETE /chapters/{id}
func DeleteChapter(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteChapter(id); err != nil {
		http.Error(w, "Failed to delete chapter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
