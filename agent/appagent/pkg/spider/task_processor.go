package spider

import (
	"fmt"
	"log"

	"github.com/go-rod/rod"
)

// ProcessURL processes a URL
func (s *HeadSpider) ProcessURL(url string) error {
	// Create a new page
	page, err := s.CreatePage()
	if err != nil {
		return fmt.Errorf("error creating page: %w", err)
	}
	defer page.Close()

	// Navigate to the URL
	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("error navigating to URL: %w", err)
	}

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("error waiting for page to load: %w", err)
	}

	// Extract session data
	if err := s.ExtractSessionData(page); err != nil {
		log.Printf("Warning: Failed to extract session data: %v", err)
	}

	// Save session data
	if err := s.SaveSessionDataToJSON(); err != nil {
		log.Printf("Warning: Failed to save session data: %v", err)
	}

	return nil
}

// ProcessBookURL processes a book URL
func (s *HeadSpider) ProcessBookURL(bookURL string, bookID string, bookHost string) error {
	log.Printf("Processing book URL: %s (ID: %s, Host: %s)", bookURL, bookID, bookHost)
	return s.ProcessURL(bookURL)
}

// ProcessChapterURL processes a chapter URL
func (s *HeadSpider) ProcessChapterURL(chapterURL string, bookID string, chapterID string, bookHost string, bookSty string) error {
	log.Printf("Processing chapter URL: %s (Book ID: %s, Chapter ID: %s, Host: %s, Style: %s)",
		chapterURL, bookID, chapterID, bookHost, bookSty)
	return s.ProcessURL(chapterURL)
}

// ProcessSessionURL processes a session URL
func (s *HeadSpider) ProcessSessionURL(url string) error {
	log.Printf("Processing session URL: %s", url)
	return s.ProcessURL(url)
}

// ProcessPageWithCallback processes a page with a callback function
func (s *HeadSpider) ProcessPageWithCallback(url string, callback func(url string, page *rod.Page, spider TaskSpider) error) error {
	// Create a new page
	page, err := s.CreatePage()
	if err != nil {
		return fmt.Errorf("error creating page: %w", err)
	}
	defer page.Close()

	// Navigate to the URL
	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("error navigating to URL: %w", err)
	}

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("error waiting for page to load: %w", err)
	}

	// Apply session data
	if err := s.LoadSessionDataFromJSON(); err != nil {
		log.Printf("Warning: Failed to load session data: %v", err)
	}

	if err := s.ApplySessionData(page); err != nil {
		log.Printf("Warning: Failed to apply session data: %v", err)
	}

	// Call the callback function
	if err := callback(url, page, s); err != nil {
		return fmt.Errorf("error in callback: %w", err)
	}

	// Extract session data
	if err := s.ExtractSessionData(page); err != nil {
		log.Printf("Warning: Failed to extract session data: %v", err)
	}

	// Save session data
	if err := s.SaveSessionDataToJSON(); err != nil {
		log.Printf("Warning: Failed to save session data: %v", err)
	}

	return nil
}
