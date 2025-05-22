package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cct/models"
)

// GetWebsites handles GET /websites
func GetWebsites(w http.ResponseWriter, r *http.Request) {
	websites, err := models.GetWebsites()
	if err != nil {
		http.Error(w, "Failed to get websites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(websites)
}

// GetWebsite handles GET /websites/{id}
func GetWebsite(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid website ID", http.StatusBadRequest)
		return
	}

	website, err := models.GetWebsite(id)
	if err != nil {
		http.Error(w, "Failed to get website: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(website)
}

// CreateWebsite handles POST /websites
func CreateWebsite(w http.ResponseWriter, r *http.Request) {
	var website models.Website
	if err := json.NewDecoder(r.Body).Decode(&website); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := models.CreateWebsite(&website); err != nil {
		http.Error(w, "Failed to create website: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(website)
}

// UpdateWebsite handles PUT /websites/{id}
func UpdateWebsite(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid website ID", http.StatusBadRequest)
		return
	}

	var website models.Website
	if err := json.NewDecoder(r.Body).Decode(&website); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	website.ID = id

	if err := models.UpdateWebsite(&website); err != nil {
		http.Error(w, "Failed to update website: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(website)
}

// DeleteWebsite handles DELETE /websites/{id}
func DeleteWebsite(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid website ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteWebsite(id); err != nil {
		http.Error(w, "Failed to delete website: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
