package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type Website struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	URL           string `json:"base_url"`
	ScriptName    string `json:"script_name"`
	CrawlInterval int    `json:"crawl_interval"`
	Enabled       bool   `json:"enabled"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

type IWebsiteService interface {
	GetWebsites(context.Context) ([]Website, error)
	GetWebsite(context.Context, int) (Website, error)
}

type WebsiteService struct {
	client *Client
}

func NewWebsiteService(client *Client) IWebsiteService {
	return &WebsiteService{
		client: client,
	}
}

func (s *WebsiteService) GetWebsites(ctx context.Context) ([]Website, error) {
	s.client.SetHeader("Content-Type", "application/json")
	resp, err := s.client.Get(ctx, "/api/websites")
	if err != nil {
		return []Website{}, fmt.Errorf("failed to get websites: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return []Website{}, fmt.Errorf("failed to get websites: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Decode the response
	var websites []Website
	if err := json.NewDecoder(resp.Body).Decode(&websites); err != nil {
		return []Website{}, fmt.Errorf("failed to decode website response: %w", err)
	}
	return websites, nil
}

func (s *WebsiteService) GetWebsite(ctx context.Context, id int) (Website, error) {
	s.client.SetHeader("Content-Type", "application/json")
	resp, err := s.client.Get(ctx, fmt.Sprintf("/api/websites/%d", id))
	if err != nil {
		return Website{}, fmt.Errorf("failed to get website: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return Website{}, fmt.Errorf("failed to get website: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Decode the response
	var website Website
	if err := json.NewDecoder(resp.Body).Decode(&website); err != nil {
		return Website{}, fmt.Errorf("failed to decode website response: %w", err)
	}
	return website, nil
}
