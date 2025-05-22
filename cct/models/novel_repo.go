package models

import (
	"database/sql"
	"fmt"

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
			&n.ID, &n.WebsiteID, &n.ExternalID, &n.Title, &n.Author, &n.Genres,
			&n.CoverURL, &n.Description, &n.Status, &n.SourceURL, &n.LastCrawledAt, &n.CreatedAt,
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
		SELECT id, website_id, external_id, title, author, genres, cover_url, 
		       description, status, source_url, last_crawled_at, created_at
		FROM novels
		WHERE id = $1
	`, id).Scan(
		&n.ID, &n.WebsiteID, &n.ExternalID, &n.Title, &n.Author, &n.Genres,
		&n.CoverURL, &n.Description, &n.Status, &n.SourceURL, &n.LastCrawledAt, &n.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Novel{}, fmt.Errorf("novel with ID %d not found", id)
		}
		return Novel{}, fmt.Errorf("failed to query novel: %w", err)
	}

	return n, nil
}

// GetNovelsByWebsite retrieves all novels for a specific website
func GetNovelsByWebsite(websiteID int) ([]Novel, error) {
	rows, err := utils.DB.Query(`
		SELECT id, website_id, external_id, title, author, genres, cover_url, 
		       description, status, source_url, last_crawled_at, created_at
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
			&n.ID, &n.WebsiteID, &n.ExternalID, &n.Title, &n.Author, &n.Genres,
			&n.CoverURL, &n.Description, &n.Status, &n.SourceURL, &n.LastCrawledAt, &n.CreatedAt,
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
		INSERT INTO novels (website_id, external_id, title, author, genres, cover_url, 
		                   description, status, source_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`, n.WebsiteID, n.ExternalID, n.Title, n.Author, n.Genres, n.CoverURL,
		n.Description, n.Status, n.SourceURL).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create novel: %w", err)
	}

	return nil
}

// UpdateNovel updates an existing novel
func UpdateNovel(n *Novel) error {
	_, err := utils.DB.Exec(`
		UPDATE novels
		SET website_id = $1, external_id = $2, title = $3, author = $4, 
		    genres = $5, cover_url = $6, description = $7, status = $8, 
		    source_url = $9, last_crawled_at = $10
		WHERE id = $11
	`, n.WebsiteID, n.ExternalID, n.Title, n.Author, n.Genres, n.CoverURL,
		n.Description, n.Status, n.SourceURL, n.LastCrawledAt, n.ID)
	if err != nil {
		return fmt.Errorf("failed to update novel: %w", err)
	}

	return nil
}

// DeleteNovel deletes a novel by ID
func DeleteNovel(id int) error {
	_, err := utils.DB.Exec("DELETE FROM novels WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete novel: %w", err)
	}
	return nil
}
