package spider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	SessionPrefix     string
}

func (s *HeadSpider) initBrowser() error {
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

func (s *HeadSpider) closeBrowser() {
	if s.browser != nil {
		s.browser.MustClose()
		s.browser = nil
	}
	if s.browserLauncher != nil {
		s.browserLauncher.Cleanup()
		s.browserLauncher = nil
	}
}

func (s *HeadSpider) fetch(ctx context.Context, rawURL string) error {
	requiresHead, err := s.requiresHeadBrowser(rawURL)
	if err != nil {
		return err
	}

	if requiresHead {
		return s.fetchWithHeadBrowser(ctx, rawURL)
	}
	return s.BasicSpider.fetch(ctx, rawURL)
}

func (s *HeadSpider) requiresHeadBrowser(rawURL string) (bool, error) {
	req, err := http.NewRequest("HEAD", rawURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return true, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 503 {
		return true, nil
	}

	securityHeaders := []string{
		"cf-ray",
		"server-timing",
		"x-robots-tag",
	}

	for _, header := range securityHeaders {
		if resp.Header.Get(header) != "" {
			return true, nil
		}
	}

	return false, nil
}

func (s *HeadSpider) fetchWithHeadBrowser(ctx context.Context, rawURL string) error {
	pageUrl := rawURL
	if strings.HasPrefix(rawURL, s.SessionPrefix) {
		s.SetHeadless(false)
		pageUrl = rawURL[len(s.SessionPrefix):]
	} else {
		if !s.isHeadless {
			s.SetHeadless(true)
		}
	}

	page := s.browser.MustPage()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	pageCtx, cancel := context.WithTimeout(ctx, s.browserTimeout)
	defer cancel()

	if len(s.cookies) > 0 {
		parsedURL, err := url.Parse(pageUrl)
		if err != nil {
			return err
		}

		domain := parsedURL.Hostname()
		for _, cookie := range s.cookies {
			if strings.HasSuffix(domain, cookie.Domain) || strings.HasSuffix("."+domain, cookie.Domain) {
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
		}
	}

	if err := page.Context(pageCtx).Navigate(pageUrl); err != nil {
		return err
	}

	if err := page.Context(pageCtx).WaitLoad(); err != nil {
		return err
	}

	if err := page.Context(pageCtx).WaitIdle(500 * time.Millisecond); err != nil {
		return err
	}

	pageCookies, err := page.Cookies([]string{})
	if err == nil {
		s.cookies = pageCookies
	}

	if err = s.ExtractSessionData(page); err != nil {
		return err
	}

	if err = s.SaveSessionDataToJSON(); err != nil {
		return err
	}

	for _, callback := range s.responseCallbacks {
		if err := callback(rawURL, page, s); err != nil {
			fmt.Println("Error in response callback:", err)
			return err
		}
	}

	if s.lastDepth == s.maxDepth {
		return nil
	}

	links, err := s.extractLinksFromPage(page, rawURL)
	if err != nil {
		return err
	}

	for _, link := range links {
		if err := s.AddURL(link); err != nil {
			continue
		}
	}

	return nil
}

func (s *HeadSpider) extractLinksFromPage(page *rod.Page, baseURL string) ([]string, error) {
	elements, err := page.Elements("a")
	if err != nil {
		return nil, err
	}

	links := make([]string, 0)
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	for _, element := range elements {
		href, err := element.Attribute("href")
		if err != nil || href == nil {
			continue
		}

		linkURL, err := url.Parse(*href)
		if err != nil {
			continue
		}

		absoluteURL := baseURLParsed.ResolveReference(linkURL).String()
		links = append(links, absoluteURL)
	}

	return links, nil
}

// worker is a custom implementation for HeadSpider that calls HeadSpider's fetch method
func (s *HeadSpider) worker(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		default:
			queueItem, err := s.queue.Pop()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if queueItem.Depth > s.maxDepth {
				continue
			}

			s.visitedMutex.Lock()
			s.visited[queueItem.URL] = true
			s.visitedSuccess[queueItem.URL] = true
			s.visitedMutex.Unlock()

			// Call HeadSpider's fetch method instead of BasicSpider's fetch method
			if err := s.fetch(ctx, queueItem.URL); err != nil {
				s.visitedMutex.Lock()
				s.visitedSuccess[queueItem.URL] = false
				s.visitedMutex.Unlock()
				continue
			}

			time.Sleep(s.delay)
		}
	}
}

func NewHeadSpider(isHeadless bool, conf *config.Config) *HeadSpider {
	return &HeadSpider{
		browserTimeout: conf.BrowserTimeout * time.Second, // Increased timeout to 2 minutes
		isHeadless:     isHeadless,
		captchaHandler: NewManualCaptchaHandler(),
		proxyURL:       conf.ProxyURL,
		sessionFile:    conf.SessionFile,
		SessionPrefix:  conf.SessionPrefix,
		BasicSpider: &BasicSpider{
			client: &http.Client{
				Timeout: conf.BrowserTimeout * time.Second,
			},
			concurrency:       conf.Concurrency,
			delay:             conf.Delay * time.Second,
			userAgent:         conf.UserAgent[0],
			maxDepth:          conf.MaxDepth,
			queue:             NewURLQueue(),
			visited:           make(map[string]bool),
			visitedSuccess:    make(map[string]bool),
			htmlCallbacks:     make(map[string]func(url string, element string) error),
			responseCallbacks: []func(url string, resp *http.Response) error{},
			stopCh:            make(chan struct{}),
		},
	}
}

func (s *HeadSpider) SetHeadless(isHeadless bool) {
	s.isHeadless = isHeadless

	if s.browser != nil {
		s.closeBrowser()
		s.initBrowser()
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

func (s *HeadSpider) Start(ctx context.Context) error {
	if err := s.initBrowser(); err != nil {
		return err
	}

	if err := s.ExecutePrepSteps(); err != nil {
		return err
	}

	// Start workers using HeadSpider's worker method
	for i := 0; i < s.concurrency; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}

	s.wg.Wait()
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

// QuitWorkers sends a signal to all workers to stop and waits for them to finish
func (s *HeadSpider) QuitWorkers() {
	// Signal all workers to stop
	close(s.stopCh)

	// Create a new channel for the next run if needed
	s.stopCh = make(chan struct{})
}

// IsIdle checks if the spider is idle (no URLs in queue and no active workers)
func (s *HeadSpider) IsIdle() bool {
	return s.queue.IsEmpty()
}

// Start with timeout to prevent hanging
func (s *HeadSpider) StartWithTimeout(ctx context.Context, timeout time.Duration) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Start the spider in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start(ctx)
	}()

	// Wait for either completion or timeout
	select {
	case err := <-errCh:
		return err
	case <-time.After(timeout):
		// Force quit workers if timeout occurs
		s.QuitWorkers()
		// Close the browser when timing out
		s.closeBrowser()
		return fmt.Errorf("spider execution timed out after %v", timeout)
	}
}
