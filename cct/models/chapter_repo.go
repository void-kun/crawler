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
		       content, crawled_at, error
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
			&c.Content, &c.CrawledAt, &c.Error,
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
		       content, crawled_at, error
		FROM chapters
		WHERE id = $1
	`, id).Scan(
		&c.ID, &c.NovelID, &c.ExternalID, &c.Title, &c.ChapterNumber, &c.URL,
		&c.Content, &c.CrawledAt, &c.Error,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Chapter{}, fmt.Errorf("chapter with ID %d not found", id)
		}
		return Chapter{}, fmt.Errorf("failed to query chapter: %w", err)
	}

	return c, nil
}

// GetChapterByUrl retrieves a chapter by URL
func GetChapterByUrl(url string) (Chapter, error) {
	var c Chapter
	err := utils.DB.QueryRow(`
		SELECT id, novel_id, external_id, title, chapter_number, url, content, crawled_at, error
		FROM chapters
		WHERE url = $1
	`, url).Scan(&c.ID, &c.NovelID, &c.ExternalID, &c.Title, &c.ChapterNumber, &c.URL, &c.Content, &c.CrawledAt, &c.Error)
	if err != nil {
		if err == sql.ErrNoRows {
			return Chapter{}, fmt.Errorf("chapter with URL %s not found", url)
		}
		return Chapter{}, fmt.Errorf("failed to query chapter: %w", err)
	}

	return c, nil
}

// GetChaptersByNovel retrieves all chapters for a specific novel
func GetChaptersByNovel(novelID int) ([]Chapter, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, external_id, title, chapter_number, url, 
		       content, crawled_at, error
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
			&c.Content, &c.CrawledAt, &c.Error,
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
		                     content)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, c.NovelID, c.ExternalID, c.Title, c.ChapterNumber, c.URL,
		c.Content).Scan(&c.ID)
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
		    url = $5, content = $6, crawled_at = $7, error = $8
		WHERE id = $9
	`, c.NovelID, c.ExternalID, c.Title, c.ChapterNumber, c.URL,
		c.Content, c.CrawledAt, c.Error, c.ID)
	if err != nil {
		return fmt.Errorf("failed to update chapter: %w", err)
	}

	return nil
}

func UpdateChapterByUrl(url string, content string) error {
	_, err := utils.DB.Exec("UPDATE chapters SET content = $1, crawled_at = now() WHERE url = $2", content, url)
	return err
}

func InsertOrUpdateChapter(c *Chapter) error {
	var existingChapter Chapter
	err := utils.DB.QueryRow("SELECT * FROM chapter WHERE novel_id = $1 AND external_id = $2", c.NovelID, c.ExternalID).Scan(&existingChapter)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = utils.DB.Exec("INSERT INTO chapter (novel_id, external_id, chapter_number, title, url, content, crawled_at) VALUES ($1, $2, $3, $4, $5, $6, $7)", c.NovelID, c.ExternalID, c.ChapterNumber, c.Title, c.URL, c.Content, c.CrawledAt)
			return err
		}
		return err
	}

	_, err = utils.DB.Exec("UPDATE chapter SET title = $1, url = $2, content = $3, crawled_at = $4 WHERE novel_id = $5 AND chapter_number = $6", c.Title, c.URL, c.Content, c.CrawledAt, c.NovelID, c.ChapterNumber)
	return err
}

// DeleteChapter deletes a chapter by ID
func DeleteChapter(id int) error {
	_, err := utils.DB.Exec("DELETE FROM chapters WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}
	return nil
}
