package models

import (
	"database/sql"
	"fmt"

	"cct/utils"
)

// GetChapters retrieves all chapters from the database
func GetChapters() ([]Chapter, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, external_id, title, chapter_number, url, 
		       content, status, crawled_at, error
		FROM chapters
		ORDER BY novel_id, chapter_number
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapters: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var c Chapter
		if err := rows.Scan(
			&c.ID, &c.NovelID, &c.ExternalID, &c.Title, &c.ChapterNumber, &c.URL,
			&c.Content, &c.Status, &c.CrawledAt, &c.Error,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chapter row: %w", err)
		}
		chapters = append(chapters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chapter rows: %w", err)
	}

	return chapters, nil
}

// GetChapter retrieves a chapter by ID
func GetChapter(id int) (Chapter, error) {
	var c Chapter
	err := utils.DB.QueryRow(`
		SELECT id, novel_id, external_id, title, chapter_number, url, 
		       content, status, crawled_at, error
		FROM chapters
		WHERE id = $1
	`, id).Scan(
		&c.ID, &c.NovelID, &c.ExternalID, &c.Title, &c.ChapterNumber, &c.URL,
		&c.Content, &c.Status, &c.CrawledAt, &c.Error,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return Chapter{}, fmt.Errorf("chapter with ID %d not found", id)
		}
		return Chapter{}, fmt.Errorf("failed to query chapter: %w", err)
	}
	
	return c, nil
}

// GetChaptersByNovel retrieves all chapters for a specific novel
func GetChaptersByNovel(novelID int) ([]Chapter, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, external_id, title, chapter_number, url, 
		       content, status, crawled_at, error
		FROM chapters
		WHERE novel_id = $1
		ORDER BY chapter_number
	`, novelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapters by novel: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var c Chapter
		if err := rows.Scan(
			&c.ID, &c.NovelID, &c.ExternalID, &c.Title, &c.ChapterNumber, &c.URL,
			&c.Content, &c.Status, &c.CrawledAt, &c.Error,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chapter row: %w", err)
		}
		chapters = append(chapters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chapter rows: %w", err)
	}

	return chapters, nil
}

// CreateChapter creates a new chapter in the database
func CreateChapter(c *Chapter) error {
	err := utils.DB.QueryRow(`
		INSERT INTO chapters (novel_id, external_id, title, chapter_number, url, 
		                     content, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, c.NovelID, c.ExternalID, c.Title, c.ChapterNumber, c.URL,
		c.Content, c.Status).Scan(&c.ID)
	
	if err != nil {
		return fmt.Errorf("failed to create chapter: %w", err)
	}
	
	return nil
}

// UpdateChapter updates an existing chapter
func UpdateChapter(c *Chapter) error {
	_, err := utils.DB.Exec(`
		UPDATE chapters
		SET novel_id = $1, external_id = $2, title = $3, chapter_number = $4, 
		    url = $5, content = $6, status = $7, crawled_at = $8, error = $9
		WHERE id = $10
	`, c.NovelID, c.ExternalID, c.Title, c.ChapterNumber, c.URL,
		c.Content, c.Status, c.CrawledAt, c.Error, c.ID)
	
	if err != nil {
		return fmt.Errorf("failed to update chapter: %w", err)
	}
	
	return nil
}

// DeleteChapter deletes a chapter by ID
func DeleteChapter(id int) error {
	_, err := utils.DB.Exec("DELETE FROM chapters WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}
	return nil
}
