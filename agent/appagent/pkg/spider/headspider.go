package spider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/zrik/agent/appagent/pkg/config"
)

type HeadSpider struct {
	*BasicSpider
	browserPath       string
	browserTimeout    time.Duration
	proxyURL          string
	browser           *rod.Browser
	browserLauncher   *launcher.Launcher
	prepSteps         []func(*rod.Browser, *HeadSpider) error
	responseCallbacks []func(url string, page *rod.Page, hs *HeadSpider) error
	cookies           []*proto.NetworkCookie
	isHeadless        bool
	captchaHandler    CaptchaHandler
	sessionData       *SessionData
	sessionFile       string
}

// CreatePage creates a new page
func (s *HeadSpider) CreatePage() (*rod.Page, error) {
	if s.browser == nil {
		if err := s.InitBrowser(); err != nil {
			return nil, err
		}
	}

	page, err := s.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (s *HeadSpider) InitBrowser() error {
	if s.browser != nil {
		return nil
	}

	s.browserLauncher = launcher.New()

	if s.browserPath != "" {
		s.browserLauncher.Bin(s.browserPath)
	}

	if s.proxyURL != "" {
		s.browserLauncher.Proxy(s.proxyURL)
	}

	// Add browser flags to improve stability and performance
	s.browserLauncher.Set("disable-web-security")
	s.browserLauncher.Set("disable-background-networking")
	s.browserLauncher.Set("disable-background-timer-throttling")
	s.browserLauncher.Set("disable-backgrounding-occluded-windows")
	s.browserLauncher.Set("disable-breakpad")
	s.browserLauncher.Set("disable-component-extensions-with-background-pages")
	s.browserLauncher.Set("disable-extensions")
	s.browserLauncher.Set("disable-features", "TranslateUI,BlinkGenPropertyTrees")
	s.browserLauncher.Set("disable-ipc-flooding-protection")
	s.browserLauncher.Set("disable-renderer-backgrounding")
	s.browserLauncher.Set("enable-features", "NetworkService,NetworkServiceInProcess")

	s.browserLauncher.Headless(s.isHeadless)
	url := s.browserLauncher.MustLaunch()
	s.browser = rod.New().ControlURL(url).MustConnect()

	return nil
}

func (s *HeadSpider) CloseBrowser() {
	if s.browser != nil {
		s.browser.MustClose()
		s.browser = nil
	}
	if s.browserLauncher != nil {
		s.browserLauncher.Cleanup()
		s.browserLauncher = nil
	}
}

func NewHeadSpider(isHeadless bool, conf *config.Config) *HeadSpider {
	return &HeadSpider{
		browserTimeout: conf.BrowserTimeout * time.Second, // Increased timeout to 2 minutes
		isHeadless:     isHeadless,
		captchaHandler: NewManualCaptchaHandler(),
		proxyURL:       conf.ProxyURL,
		sessionFile:    conf.SessionFile,
		BasicSpider: &BasicSpider{
			client: &http.Client{
				Timeout: conf.BrowserTimeout * time.Second,
			},
			concurrency:       conf.Concurrency,
			delay:             conf.Delay * time.Second,
			userAgent:         conf.UserAgent[0],
			queue:             NewURLQueue(),
			visited:           make(map[string]bool),
			visitedSuccess:    make(map[string]bool),
			htmlCallbacks:     make(map[string]func(url string, element string) error),
			responseCallbacks: []func(url string, resp *http.Response) error{},
		},
	}
}

func (s *HeadSpider) SetHeadless(isHeadless bool) {
	s.isHeadless = isHeadless

	if s.browser != nil {
		s.CloseBrowser()
		s.InitBrowser()
	}
}

func (s *HeadSpider) SetBrowserPath(path string) {
	s.browserPath = path
}

func (s *HeadSpider) SetBrowserTimeout(timeout time.Duration) {
	s.browserTimeout = timeout
}

func (s *HeadSpider) SetProxy(proxyURL string) {
	s.proxyURL = proxyURL
}

func (s *HeadSpider) SetCaptchaHandler(handler CaptchaHandler) {
	s.captchaHandler = handler
}

func (s *HeadSpider) HandleCaptcha(page *rod.Page) error {
	if s.captchaHandler == nil {
		// If no handler is set, use the default manual handler
		s.captchaHandler = NewManualCaptchaHandler()
	}

	// Check if a captcha is present
	if DetectCaptcha(page) {
		return s.captchaHandler.HandleCaptcha(page)
	}

	return nil
}

func (s *HeadSpider) AddPrepStep(step func(*rod.Browser, *HeadSpider) error) {
	s.prepSteps = append(s.prepSteps, step)
}

func (s *HeadSpider) ExecutePreStep(step func(*rod.Browser, *HeadSpider) error) error {
	if s.browser == nil {
		return fmt.Errorf("browser not initialized")
	}

	if err := step(s.browser, s); err != nil {
		return err
	}

	return nil
}

func (s *HeadSpider) ExecutePrepSteps() error {
	for _, step := range s.prepSteps {
		if err := step(s.browser, s); err != nil {
			return err
		}
	}

	return nil
}

func (s *HeadSpider) GetCookies() []*proto.NetworkCookie {
	return s.cookies
}

func (s *HeadSpider) SetCookies(cookies []*proto.NetworkCookie) {
	s.cookies = cookies
}

func (s *HeadSpider) LoadCookiesFromJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return err
	}

	s.cookies = cookies
	return nil
}

func (s *HeadSpider) SaveCookiesToJSON(filePath string) error {
	data, err := json.MarshalIndent(s.cookies, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

// ExtractSessionData extracts all session data from the current page
func (s *HeadSpider) ExtractSessionData(page *rod.Page) error {
	sessionData, err := ExtractSessionData(page)
	if err != nil {
		return err
	}

	s.sessionData = sessionData
	return nil
}

// GetSessionData returns the current session data
func (s *HeadSpider) GetSessionData() *SessionData {
	return s.sessionData
}

// SaveSessionDataToJSON saves the current session data to a JSON file
func (s *HeadSpider) SaveSessionDataToJSON() error {
	if s.sessionData == nil {
		return fmt.Errorf("no session data available")
	}

	return SaveSessionDataToJSON(s.sessionData, s.sessionFile)
}

// LoadSessionDataFromJSON loads session data from a JSON file
func (s *HeadSpider) LoadSessionDataFromJSON() error {
	sessionData, err := LoadSessionDataFromJSON(s.sessionFile)
	if err != nil {
		return err
	}

	s.sessionData = sessionData
	s.cookies = sessionData.Cookies

	return nil
}

// ApplySessionData applies the current session data to a page
func (s *HeadSpider) ApplySessionData(page *rod.Page) error {
	if s.sessionData == nil {
		return fmt.Errorf("no session data available")
	}

	return ApplySessionDataToPage(page, s.sessionData)
}

func (s *HeadSpider) OnResponse(callback func(url string, page *rod.Page, hs *HeadSpider) error) {
	s.responseCallbacks = append(s.responseCallbacks, callback)
}

// IsIdle checks if the spider is idle (no URLs in queue and no active workers)
func (s *HeadSpider) IsIdle() bool {
	return s.queue.IsEmpty()
}
