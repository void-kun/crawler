package models

import (
	"database/sql"
	"fmt"
	"time"

	"cct/utils"
)

// GetSchedules retrieves all novel schedules from the database
func GetSchedules() ([]NovelSchedule, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, enabled, interval_seconds, last_run_at, next_run_at, created_at, updated_at
		FROM novel_schedules
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []NovelSchedule
	for rows.Next() {
		var s NovelSchedule
		if err := rows.Scan(
			&s.ID, &s.NovelID, &s.Enabled, &s.IntervalSeconds, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan schedule row: %w", err)
		}
		schedules = append(schedules, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	return schedules, nil
}

// GetSchedule retrieves a schedule by ID
func GetSchedule(id int) (NovelSchedule, error) {
	var s NovelSchedule
	err := utils.DB.QueryRow(`
		SELECT id, novel_id, enabled, interval_seconds, last_run_at, next_run_at, created_at, updated_at
		FROM novel_schedules
		WHERE id = $1
	`, id).Scan(
		&s.ID, &s.NovelID, &s.Enabled, &s.IntervalSeconds, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return NovelSchedule{}, fmt.Errorf("schedule with ID %d not found", id)
		}
		return NovelSchedule{}, fmt.Errorf("failed to query schedule: %w", err)
	}

	return s, nil
}

// GetSchedulesByNovel retrieves schedules for a specific novel
func GetSchedulesByNovel(novelID int) ([]NovelSchedule, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, enabled, interval_seconds, last_run_at, next_run_at, created_at, updated_at
		FROM novel_schedules
		WHERE novel_id = $1
		ORDER BY id
	`, novelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules for novel %d: %w", novelID, err)
	}
	defer rows.Close()

	var schedules []NovelSchedule
	for rows.Next() {
		var s NovelSchedule
		if err := rows.Scan(
			&s.ID, &s.NovelID, &s.Enabled, &s.IntervalSeconds, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan schedule row: %w", err)
		}
		schedules = append(schedules, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	return schedules, nil
}

// CreateSchedule creates a new schedule in the database
func CreateSchedule(s *NovelSchedule) error {
	// Set next_run_at to now + interval if not provided
	if !s.NextRunAt.Valid {
		nextRun := time.Now().Add(time.Duration(s.IntervalSeconds) * time.Second)
		s.NextRunAt = sql.NullTime{Time: nextRun, Valid: true}
	}

	err := utils.DB.QueryRow(`
		INSERT INTO novel_schedules (novel_id, enabled, interval_seconds, next_run_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, s.NovelID, s.Enabled, s.IntervalSeconds, s.NextRunAt).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return nil
}

// UpdateSchedule updates an existing schedule
func UpdateSchedule(s *NovelSchedule) error {
	s.UpdatedAt = time.Now()
	_, err := utils.DB.Exec(`
		UPDATE novel_schedules
		SET enabled = $1, interval_seconds = $2, next_run_at = $3, updated_at = $4
		WHERE id = $5
	`, s.Enabled, s.IntervalSeconds, s.NextRunAt, s.UpdatedAt, s.ID)
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return nil
}

// DeleteSchedule deletes a schedule by ID
func DeleteSchedule(id int) error {
	_, err := utils.DB.Exec("DELETE FROM novel_schedules WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}
	return nil
}

// GetDueSchedules retrieves schedules that are due to run
func GetDueSchedules() ([]NovelSchedule, error) {
	rows, err := utils.DB.Query(`
		SELECT id, novel_id, enabled, interval_seconds, last_run_at, next_run_at, created_at, updated_at
		FROM novel_schedules
		WHERE enabled = true AND next_run_at IS NOT NULL AND next_run_at <= NOW()
		ORDER BY next_run_at
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query due schedules: %w", err)
	}
	defer rows.Close()

	var schedules []NovelSchedule
	for rows.Next() {
		var s NovelSchedule
		if err := rows.Scan(
			&s.ID, &s.NovelID, &s.Enabled, &s.IntervalSeconds, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan schedule row: %w", err)
		}
		schedules = append(schedules, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	return schedules, nil
}

// UpdateScheduleRunTime updates the last_run_at and next_run_at for a schedule
func UpdateScheduleRunTime(id int, intervalSeconds int) error {
	now := time.Now()
	nextRun := now.Add(time.Duration(intervalSeconds) * time.Second)

	_, err := utils.DB.Exec(`
		UPDATE novel_schedules
		SET last_run_at = $1, next_run_at = $2, updated_at = $3
		WHERE id = $4
	`, now, nextRun, now, id)
	if err != nil {
		return fmt.Errorf("failed to update schedule run time: %w", err)
	}

	return nil
}

// TriggerScheduleNow updates the next_run_at to current time to trigger immediate execution
func TriggerScheduleNow(id int) error {
	now := time.Now()

	_, err := utils.DB.Exec(`
		UPDATE novel_schedules
		SET next_run_at = $1, updated_at = $2
		WHERE id = $3
	`, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to trigger schedule: %w", err)
	}

	return nil
}

// UpdateNovelLastCrawledAt updates the last_crawled_at timestamp for a novel
func UpdateNovelLastCrawledAt(novelID int) error {
	_, err := utils.DB.Exec(`
		UPDATE novels
		SET last_crawled_at = NOW()
		WHERE id = $1
	`, novelID)
	if err != nil {
		return fmt.Errorf("failed to update novel last_crawled_at: %w", err)
	}

	return nil
}

// CreateChapterCrawlLog creates a log entry for chapter crawling
func CreateChapterCrawlLog(log *ChapterCrawlLog) error {
	err := utils.DB.QueryRow(`
		INSERT INTO chapter_crawl_logs (chapter_id, status, error)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, log.ChapterID, log.Status, log.Error).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create chapter crawl log: %w", err)
	}

	return nil
}

// GetChapterCrawlLogs retrieves crawl logs for a chapter
func GetChapterCrawlLogs(chapterID int) ([]ChapterCrawlLog, error) {
	rows, err := utils.DB.Query(`
		SELECT id, chapter_id, status, error, created_at
		FROM chapter_crawl_logs
		WHERE chapter_id = $1
		ORDER BY created_at DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapter crawl logs: %w", err)
	}
	defer rows.Close()

	var logs []ChapterCrawlLog
	for rows.Next() {
		var log ChapterCrawlLog
		if err := rows.Scan(
			&log.ID, &log.ChapterID, &log.Status, &log.Error, &log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chapter crawl log row: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chapter crawl log rows: %w", err)
	}

	return logs, nil
}
