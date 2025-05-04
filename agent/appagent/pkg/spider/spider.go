package spider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Spider interface {
	Start(ctx context.Context, seeds []string) error
	Stop()
	AddURL(url string) error
	SetConcurrency(n int)
	SetDelay(delay time.Duration)
	SetUserAgent(ua string)
	SetMaxDepth(depth int)
	OnHTML(selector string, callback func(url string, element string) error)
	OnResponse(callback func(url string, resp *http.Response) error)
	GetQueue() []QueueItem
}

type BasicSpider struct {
	client            *http.Client
	concurrency       int
	delay             time.Duration
	userAgent         string
	maxDepth          int
	lastDepth         int
	queue             *URLQueue
	visited           map[string]bool
	visitedSuccess    map[string]bool
	visitedMutex      sync.RWMutex
	htmlCallbacks     map[string]func(url string, element string) error
	responseCallbacks []func(url string, resp *http.Response) error
	wg                sync.WaitGroup
	stopCh            chan struct{}
}

func NewBasicSpider() *BasicSpider {
	return &BasicSpider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		concurrency:       5,
		delay:             1 * time.Second,
		userAgent:         "GoSpider/1.0",
		maxDepth:          3,
		queue:             NewURLQueue(),
		visited:           make(map[string]bool),
		visitedSuccess:    make(map[string]bool),
		htmlCallbacks:     make(map[string]func(url string, element string) error),
		responseCallbacks: []func(url string, resp *http.Response) error{},
		stopCh:            make(chan struct{}),
	}
}

func (s *BasicSpider) Start(ctx context.Context, seeds []string) error {
	if len(seeds) == 0 {
		return errors.New("no seed URLs provided")
	}

	for _, seed := range seeds {
		if err := s.AddURL(seed); err != nil {
			return err
		}
	}

	for i := 0; i < s.concurrency; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}

	s.wg.Wait()
	return nil
}

func (s *BasicSpider) Stop() {
	close(s.stopCh)
}

func (s *BasicSpider) AddURL(rawURL string) error {
	if rawURL[len(rawURL)-1] != '/' {
		rawURL = rawURL + "/"
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	normalizedURL := parsedURL.String()

	s.visitedMutex.RLock()
	visited := s.visited[normalizedURL]
	s.visitedMutex.RUnlock()

	if !visited {
		s.queue.Push(normalizedURL, 0)
		s.lastDepth = 0
	}

	return nil
}

func (s *BasicSpider) SetConcurrency(n int) {
	if n > 0 {
		s.concurrency = n
	}
}

func (s *BasicSpider) SetDelay(delay time.Duration) {
	s.delay = delay
}

func (s *BasicSpider) SetUserAgent(ua string) {
	s.userAgent = ua
}

func (s *BasicSpider) SetMaxDepth(depth int) {
	if depth >= 0 {
		s.maxDepth = depth
	}
}

func (s *BasicSpider) OnHTML(selector string, callback func(url string, element string) error) {
	s.htmlCallbacks[selector] = callback
}

func (s *BasicSpider) OnResponse(callback func(url string, resp *http.Response) error) {
	s.responseCallbacks = append(s.responseCallbacks, callback)
}

func (s *BasicSpider) worker(ctx context.Context) {
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
			s.visitedMutex.Unlock()

			if err := s.fetch(ctx, queueItem.URL); err != nil {
				continue
			}

			time.Sleep(s.delay)
		}
	}
}

func (s *BasicSpider) fetch(ctx context.Context, rawURL string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", s.userAgent)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for _, callback := range s.responseCallbacks {
		if err := callback(rawURL, resp); err != nil {
			return err
		}
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (f *BasicSpider) GetQueue() []QueueItem {
	return f.queue.List()
}
