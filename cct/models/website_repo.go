package models

import (
	"database/sql"
	"fmt"

	"cct/utils"
)

// GetWebsites retrieves all websites from the database
func GetWebsites() ([]Website, error) {
	rows, err := utils.DB.Query(`
		SELECT id, name, base_url, script_name, crawl_interval, enabled, created_at
		FROM websites
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query websites: %w", err)
	}
	defer rows.Close()

	var websites []Website
	for rows.Next() {
		var w Website
		if err := rows.Scan(&w.ID, &w.Name, &w.BaseURL, &w.ScriptName, &w.CrawlInterval, &w.Enabled, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan website row: %w", err)
		}
		websites = append(websites, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating website rows: %w", err)
	}

	return websites, nil
}

// GetWebsite retrieves a website by ID
func GetWebsite(id int) (Website, error) {
	var w Website
	err := utils.DB.QueryRow(`
		SELECT id, name, base_url, script_name, crawl_interval, enabled, created_at
		FROM websites
		WHERE id = $1
	`, id).Scan(&w.ID, &w.Name, &w.BaseURL, &w.ScriptName, &w.CrawlInterval, &w.Enabled, &w.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Website{}, fmt.Errorf("website with ID %d not found", id)
		}
		return Website{}, fmt.Errorf("failed to query website: %w", err)
	}

	return w, nil
}

// CreateWebsite creates a new website in the database
func CreateWebsite(w *Website) error {
	err := utils.DB.QueryRow(`
		INSERT INTO websites (name, base_url, script_name, crawl_interval, enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, w.Name, w.BaseURL, w.ScriptName, w.CrawlInterval, w.Enabled).Scan(&w.ID, &w.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create website: %w", err)
	}

	return nil
}

// UpdateWebsite updates an existing website
func UpdateWebsite(w *Website) error {
	_, err := utils.DB.Exec(`
		UPDATE websites
		SET name = $1, base_url = $2, script_name = $3, crawl_interval = $4, enabled = $5
		WHERE id = $6
	`, w.Name, w.BaseURL, w.ScriptName, w.CrawlInterval, w.Enabled, w.ID)
	if err != nil {
		return fmt.Errorf("failed to update website: %w", err)
	}

	return nil
}

// DeleteWebsite deletes a website by ID
func DeleteWebsite(id int) error {
	_, err := utils.DB.Exec("DELETE FROM websites WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete website: %w", err)
	}
	return nil
}
