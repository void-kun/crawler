package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NullTime is a wrapper around sql.NullTime that implements JSON marshaling
type NullTime struct {
	sql.NullTime
}

// MarshalJSON implements json.Marshaler
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}

// UnmarshalJSON implements json.Unmarshaler
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	var t *time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	if t == nil {
		nt.Valid = false
		return nil
	}
	nt.Valid = true
	nt.Time = *t
	return nil
}

// Website represents a website to crawl
type Website struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	BaseURL       string    `json:"base_url"`
	ScriptName    string    `json:"script_name"`
	CrawlInterval int       `json:"crawl_interval"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
}

// Novel represents a novel from a website
type Novel struct {
	ID            int          `json:"id"`
	WebsiteID     int          `json:"website_id"`
	ExternalID    string       `json:"external_id"`
	Title         string       `json:"title"`
	Author        string       `json:"author"`
	Genres        []string     `json:"genres"`
	CoverURL      string       `json:"cover_url"`
	Description   string       `json:"description"`
	Status        string       `json:"status"` // ongoing or completed
	SourceURL     string       `json:"source_url"`
	LastCrawledAt sql.NullTime `json:"last_crawled_at"`
	CreatedAt     time.Time    `json:"created_at"`
}

// Chapter represents a chapter of a novel
type Chapter struct {
	ID            int          `json:"id"`
	NovelID       int          `json:"novel_id"`
	ExternalID    string       `json:"external_id"`
	Title         string       `json:"title"`
	ChapterNumber int          `json:"chapter_number"`
	URL           string       `json:"url"`
	Content       string       `json:"content"`
	Status        string       `json:"status"` // pending, crawled, failed
	CrawledAt     sql.NullTime `json:"crawled_at"`
	Error         string       `json:"error"`
}

// CrawlJob represents a job to crawl a novel
type CrawlJob struct {
	ID         int          `json:"id"`
	NovelID    int          `json:"novel_id"`
	Status     string       `json:"status"` // pending, in_progress, success, failed
	CreatedAt  time.Time    `json:"created_at"`
	StartedAt  sql.NullTime `json:"started_at"`
	FinishedAt sql.NullTime `json:"finished_at"`
	Error      string       `json:"error"`
}

// Agent represents a crawler agent
type Agent struct {
	ID            uuid.UUID    `json:"id"`
	Name          string       `json:"name"`
	IPAddress     string       `json:"ip_address"`
	LastHeartbeat sql.NullTime `json:"last_heartbeat"`
	IsActive      bool         `json:"is_active"`
	CreatedAt     time.Time    `json:"created_at"`
}

// User represents a user for API access control
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Don't expose password hash in JSON
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// APIToken represents an API token for authentication
type APIToken struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Token       string    `json:"token"`
	Description string    `json:"description"`
	ExpiresAt   NullTime  `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  NullTime  `json:"last_used_at"`
}
