package models

import (
	"database/sql"
	"fmt"

	"cct/utils"
)

// GetCrawlJobs retrieves all crawl jobs from the database
func GetCrawlJobs() ([]CrawlJob, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, status, created_at, started_at, finished_at, error
		FROM crawl_jobs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query crawl jobs: %w", err)
	}
	defer rows.Close()

	var jobs []CrawlJob
	for rows.Next() {
		var j CrawlJob
		if err := rows.Scan(
			&j.ID, &j.NovelID, &j.Status, &j.CreatedAt, &j.StartedAt, &j.FinishedAt, &j.Error,
		); err != nil {
			return nil, fmt.Errorf("failed to scan crawl job row: %w", err)
		}
		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating crawl job rows: %w", err)
	}

	return jobs, nil
}

// GetCrawlJob retrieves a crawl job by ID
func GetCrawlJob(id int) (CrawlJob, error) {
	var j CrawlJob
	err := utils.DB.QueryRow(`
		SELECT id, novel_id, status, created_at, started_at, finished_at, error
		FROM crawl_jobs
		WHERE id = $1
	`, id).Scan(
		&j.ID, &j.NovelID, &j.Status, &j.CreatedAt, &j.StartedAt, &j.FinishedAt, &j.Error,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return CrawlJob{}, fmt.Errorf("crawl job with ID %d not found", id)
		}
		return CrawlJob{}, fmt.Errorf("failed to query crawl job: %w", err)
	}

	return j, nil
}

// GetCrawlJobsByNovel retrieves all crawl jobs for a specific novel
func GetCrawlJobsByNovel(novelID int) ([]CrawlJob, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, status, created_at, started_at, finished_at, error
		FROM crawl_jobs
		WHERE novel_id = $1
		ORDER BY created_at DESC
	`, novelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query crawl jobs by novel: %w", err)
	}
	defer rows.Close()

	var jobs []CrawlJob
	for rows.Next() {
		var j CrawlJob
		if err := rows.Scan(
			&j.ID, &j.NovelID, &j.Status, &j.CreatedAt, &j.StartedAt, &j.FinishedAt, &j.Error,
		); err != nil {
			return nil, fmt.Errorf("failed to scan crawl job row: %w", err)
		}
		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating crawl job rows: %w", err)
	}

	return jobs, nil
}

// CreateCrawlJob creates a new crawl job in the database
func CreateCrawlJob(j *CrawlJob) error {
	err := utils.DB.QueryRow(`
		INSERT INTO crawl_jobs (novel_id, status)
		VALUES ($1, $2)
		RETURNING id, created_at
	`, j.NovelID, j.Status).Scan(&j.ID, &j.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create crawl job: %w", err)
	}

	return nil
}

// UpdateCrawlJob updates an existing crawl job
func UpdateCrawlJob(j *CrawlJob) error {
	_, err := utils.DB.Exec(`
		UPDATE crawl_jobs
		SET novel_id = $1, status = $2, started_at = $3, finished_at = $4, error = $5
		WHERE id = $6
	`, j.NovelID, j.Status, j.StartedAt, j.FinishedAt, j.Error, j.ID)
	if err != nil {
		return fmt.Errorf("failed to update crawl job: %w", err)
	}

	return nil
}

// StartCrawlJob marks a crawl job as in_progress and sets the started_at time
func StartCrawlJob(id int) error {
	_, err := utils.DB.Exec(`
		UPDATE crawl_jobs
		SET status = 'in_progress', started_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to start crawl job: %w", err)
	}

	return nil
}

// CompleteCrawlJob marks a crawl job as success and sets the finished_at time
func CompleteCrawlJob(id int) error {
	_, err := utils.DB.Exec(`
		UPDATE crawl_jobs
		SET status = 'success', finished_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to complete crawl job: %w", err)
	}

	return nil
}

// FailCrawlJob marks a crawl job as failed, sets the finished_at time, and records the error
func FailCrawlJob(id int, errorMsg string) error {
	_, err := utils.DB.Exec(`
		UPDATE crawl_jobs
		SET status = 'failed', finished_at = NOW(), error = $1
		WHERE id = $2
	`, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to fail crawl job: %w", err)
	}

	return nil
}

// DeleteCrawlJob deletes a crawl job by ID
func DeleteCrawlJob(id int) error {
	_, err := utils.DB.Exec("DELETE FROM crawl_jobs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete crawl job: %w", err)
	}
	return nil
}
