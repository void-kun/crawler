package models

import (
	"database/sql"
	"fmt"
	"strconv"

	"cct/utils"
)

// GetNovels retrieves all novels from the database
func GetNovels() ([]Novel, error) {
	rows, err := utils.DB.Query(`
		SELECT id, website_id, external_id, title, author, genres, cover_url, 
		       description, status, source_url, last_crawled_at, created_at
		FROM novels
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query novels: %w", err)
	}
	defer rows.Close()

	var novels []Novel
	for rows.Next() {
		var n Novel
		if err := rows.Scan(
			&n.ID, &n.WebsiteID, &n.ExternalID, &n.Title, &n.Status, &n.SourceURL, &n.LastCrawledAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan novel row: %w", err)
		}
		novels = append(novels, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating novel rows: %w", err)
	}

	return novels, nil
}

// GetNovel retrieves a novel by ID
func GetNovel(id int) (Novel, error) {
	var n Novel
	err := utils.DB.QueryRow(`
		SELECT id, website_id, external_id, title, status, source_url, last_crawled_at, created_at
		FROM novels
		WHERE id = $1
	`, id).Scan(
		&n.ID, &n.WebsiteID, &n.ExternalID, &n.Title, &n.Status, &n.SourceURL, &n.LastCrawledAt, &n.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Novel{}, fmt.Errorf("novel with ID %d not found", id)
		}
		return Novel{}, fmt.Errorf("failed to query novel: %w", err)
	}

	return n, nil
}

// GetNovelByUrl retrieves a novel by source URL
func GetNovelByUrl(url string) (Novel, error) {
	var n Novel
	err := utils.DB.QueryRow(`
		SELECT id, website_id, external_id, title, source_url, last_crawled_at, created_at
		FROM novels
		WHERE source_url = $1
	`, url).Scan(
		&n.ID, &n.WebsiteID, &n.ExternalID, &n.Title, &n.SourceURL, &n.LastCrawledAt, &n.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Novel{}, fmt.Errorf("novel with source URL %s not found", url)
		}
		return Novel{}, fmt.Errorf("failed to query novel: %w", err)
	}

	return n, nil
}

// GetNovelsByWebsite retrieves all novels for a specific website
func GetNovelsByWebsite(websiteID int) ([]Novel, error) {
	rows, err := utils.DB.Query(`
		SELECT id, website_id, external_id, title, status, source_url, last_crawled_at, created_at
		FROM novels
		WHERE website_id = $1
		ORDER BY id
	`, websiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query novels by website: %w", err)
	}
	defer rows.Close()

	var novels []Novel
	for rows.Next() {
		var n Novel
		if err := rows.Scan(
			&n.ID, &n.WebsiteID, &n.ExternalID, &n.Title, &n.Status, &n.SourceURL, &n.LastCrawledAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan novel row: %w", err)
		}
		novels = append(novels, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating novel rows: %w", err)
	}

	return novels, nil
}

// CreateNovel creates a new novel in the database
func CreateNovel(n *Novel) error {
	err := utils.DB.QueryRow(`
		INSERT INTO novels (website_id, title, external_id, source_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, n.WebsiteID, n.Title, n.ExternalID, n.SourceURL).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create novel: %w", err)
	}

	return nil
}

// UpdateNovel updates an existing novel
func UpdateNovel(n *Novel) error {
	var columns []string
	var values []interface{}
	if n.Title != "" {
		columns = append(columns, "title")
		values = append(values, n.Title)
	}
	if n.Status != "" {
		columns = append(columns, "status")
		values = append(values, n.Status)
	}
	if n.SourceURL != "" {
		columns = append(columns, "source_url")
		values = append(values, n.SourceURL)
	}

	updateSQL := "UPDATE novels SET "
	for i, column := range columns {
		updateSQL += column + " = $" + strconv.Itoa(i+1)
		if i < len(columns)-1 {
			updateSQL += ", "
		}
	}
	updateSQL += " WHERE id = $" + strconv.Itoa(len(columns)+1)

	_, err := utils.DB.Exec(updateSQL, append(values, n.ID)...)
	return err
}

// DeleteNovel deletes a novel by ID
func DeleteNovel(id int) error {
	_, err := utils.DB.Exec("DELETE FROM novels WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete novel: %w", err)
	}
	return nil
}
