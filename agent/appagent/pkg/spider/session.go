package spider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// SessionData represents all session-related data from a browser
type SessionData struct {
	URL            string                 `json:"url"`
	Headers        map[string]string      `json:"headers"`
	Cookies        []*proto.NetworkCookie `json:"cookies"`
	LocalStorage   map[string]string      `json:"local_storage"`
	SessionStorage map[string]string      `json:"session_storage"`
	Timestamp      time.Time              `json:"timestamp"`
}

// NewSessionData creates a new empty SessionData object
func NewSessionData(url string) *SessionData {
	return &SessionData{
		URL:            url,
		Headers:        make(map[string]string),
		Cookies:        []*proto.NetworkCookie{},
		LocalStorage:   make(map[string]string),
		SessionStorage: make(map[string]string),
		Timestamp:      time.Now(),
	}
}

// ExtractSessionData extracts all session data from a page
func ExtractSessionData(page *rod.Page) (*SessionData, error) {
	url := page.MustInfo().URL

	// Create a new session data object
	sessionData := NewSessionData(url)

	// Extract cookies
	cookies, err := page.Cookies([]string{})
	if err == nil {
		sessionData.Cookies = cookies
	} else {
		fmt.Printf("Error extracting cookies: %v\n", err)
	}

	// Extract localStorage
	localStorage, err := extractLocalStorage(page)
	if err == nil {
		sessionData.LocalStorage = localStorage
	} else {
		fmt.Printf("Error extracting localStorage: %v\n", err)
	}

	// Extract sessionStorage
	sessionStorage, err := extractSessionStorage(page)
	if err == nil {
		sessionData.SessionStorage = sessionStorage
	} else {
		fmt.Printf("Error extracting sessionStorage: %v\n", err)
	}

	return sessionData, nil
}

// extractLocalStorage extracts all localStorage items from a page
func extractLocalStorage(page *rod.Page) (map[string]string, error) {
	result := make(map[string]string)

	// Use JavaScript to get all localStorage items
	jsResult, err := page.Eval(`() => {
		const items = {};
		for (let i = 0; i < localStorage.length; i++) {
			const key = localStorage.key(i);
			items[key] = localStorage.getItem(key);
		}
		return JSON.stringify(items);
	}`)
	if err != nil {
		return result, err
	}

	// Convert the JavaScript result to a Go map
	jsonStr := jsResult.Value.String()

	// Unmarshal the JSON string into a map
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// extractSessionStorage extracts all sessionStorage items from a page
func extractSessionStorage(page *rod.Page) (map[string]string, error) {
	result := make(map[string]string)

	// Use JavaScript to get all sessionStorage items
	jsResult, err := page.Eval(`() => {
		const items = {};
		for (let i = 0; i < sessionStorage.length; i++) {
			const key = sessionStorage.key(i);
			items[key] = sessionStorage.getItem(key);
		}
		return JSON.stringify(items);
	}`)
	if err != nil {
		return result, err
	}

	// Convert the JavaScript result to a Go map
	jsonStr := jsResult.Value.String()

	// Unmarshal the JSON string into a map
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// SaveSessionDataToJSON saves session data to a JSON file
func SaveSessionDataToJSON(sessionData *SessionData, filePath string) error {
	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

// LoadSessionDataFromJSON loads session data from a JSON file
func LoadSessionDataFromJSON(filePath string) (*SessionData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, err
	}

	return &sessionData, nil
}

// ApplySessionDataToPage applies session data to a page
func ApplySessionDataToPage(page *rod.Page, sessionData *SessionData) error {
	// Apply cookies
	for _, cookie := range sessionData.Cookies {
		// Using MustSetCookies which doesn't return an error
		page.MustSetCookies(&proto.NetworkCookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
			SameSite: cookie.SameSite,
		})
	}

	// Apply localStorage
	if len(sessionData.LocalStorage) > 0 {
		script := "() => {\n"
		for key, value := range sessionData.LocalStorage {
			script += fmt.Sprintf("  localStorage.setItem(%q, %q);\n", key, value)
		}
		script += "  return true;\n};"

		_, err := page.Eval(script)
		if err != nil {
			return err
		}
	}

	// Apply sessionStorage
	if len(sessionData.SessionStorage) > 0 {
		script := "() => {\n"
		for key, value := range sessionData.SessionStorage {
			script += fmt.Sprintf("  sessionStorage.setItem(%q, %q);\n", key, value)
		}
		script += "  return true;\n};"

		_, err := page.Eval(script)
		if err != nil {
			return err
		}
	}

	return nil
}

// ConvertHeadersToMap converts http.Header to a map[string]string
func ConvertHeadersToMap(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		result[key] = values[0]
		if len(values) > 1 {
			for i := 1; i < len(values); i++ {
				result[key] += ", " + values[i]
			}
		}
	}
	return result
}
