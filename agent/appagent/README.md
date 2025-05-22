# Go Spider Engine

A flexible web crawler for both normal and secure websites, written in Go.

## Features

- Crawl normal websites with HTTP requests
- Handle secure websites with JavaScript challenges using a headless browser (Rod)
- Automatically detect websites that require JavaScript rendering
- Manual captcha handling with user intervention
- Language selection support for international websites
- Session data extraction and reuse (cookies, localStorage, sessionStorage)
- Configurable concurrency, delay, and depth
- Save crawled data to files
- Command-line interface for easy use
- RabbitMQ integration for distributed task processing

## Requirements

- Go 1.16 or higher
- Chrome or Chromium (for headless browser support)
  - Rod will automatically download a browser if none is specified

## Installation

```bash
# Clone the repository
git clone https://github.com/zrik/crawler.git
cd crawler

# Build the binary
go build -o spider
```

## Usage

### Basic Usage

```bash
# Crawl a single website
./spider -seeds="https://example.com"

# Crawl multiple websites
./spider -seeds="https://example.com,https://example.org"
```

### Advanced Options

```bash
# Use headless browser for JavaScript challenges
./spider -seeds="https://example.com" -headless

# Set concurrency and delay
./spider -seeds="https://example.com" -concurrency=10 -delay=2s

# Set maximum crawl depth
./spider -seeds="https://example.com" -depth=5

# Specify output directory
./spider -seeds="https://example.com" -output="./crawled_data"
```

### Configuration File

You can also use a JSON configuration file:

```bash
./spider -config="config.json"
```

Example `config.json`:

```json
{
  "seeds": ["https://example.com", "https://example.org"],
  "concurrency": 10,
  "delay": "2s",
  "user_agent": "GoSpider/1.0",
  "max_depth": 5,
  "use_headless": true,
  "browser_path": "/usr/bin/chromium",
  "browser_timeout": "60s",
  "output_dir": "./crawled_data"
}
```

## Implementation Details

### Spider Interface

The spider engine is built around the `Spider` interface, which defines the core functionality:

```go
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
}
```

### Basic Spider

The `BasicSpider` implementation handles normal websites using standard HTTP requests.

### Headless Spider

The `HeadSpider` extends the `BasicSpider` to handle secure websites with JavaScript challenges using a headless browser. It uses the [Rod](https://github.com/go-rod/rod) library for browser automation, which provides:

- Automatic browser management (download, launch, cleanup)
- JavaScript execution and rendering
- Waiting for page load and network idle
- Element selection and interaction
- Handling of complex JavaScript challenges
- Captcha detection and handling

#### Captcha Handling

The spider includes a captcha handling system that can:

- Detect various types of captchas (reCAPTCHA, hCaptcha, etc.)
- Take screenshots to help with manual solving
- Wait for user intervention to solve captchas
- Continue crawling after captcha resolution

See the `examples/manual_captcha` directory for a demonstration of captcha handling.

#### Session Data Management

The spider includes a comprehensive session data management system that can:

- Extract and save all session-related data:
  - Cookies
  - localStorage
  - sessionStorage
- Save session data to JSON files
- Load session data from JSON files
- Apply session data to new browser sessions
- Reuse login sessions without re-authentication

See the `examples/session_reuse` directory for a demonstration of session data reuse.

## RabbitMQ Integration

The spider engine includes RabbitMQ integration for distributed task processing. This allows you to:

- Receive tasks from RabbitMQ queues
- Process tasks based on their topic
- Prioritize specific types of tasks
- Distribute crawling tasks across multiple instances

### Running the RabbitMQ Worker

```bash
# Build the worker
make build-worker

# Run the worker
make run-worker
```

### Publishing Tasks

```bash
# Build the publisher
make build-publisher

# Publish a session task
make run-publisher-session URL="https://sangtacviet.app/truyen/12345"

# Publish a book task
make run-publisher-book URL="https://sangtacviet.app/truyen/12345" BOOK_ID="12345" BOOK_HOST="sangtacviet"

# Publish a chapter task
make run-publisher-chapter URL="https://sangtacviet.app/truyen/12345/67890" BOOK_ID="12345" CHAPTER_ID="67890" BOOK_HOST="sangtacviet" BOOK_STY="truyen"
```

See the [RabbitMQ Integration README](pkg/rabbitmq/README.md) for more details.

## TODO

- Improve HTML parsing and link extraction for the basic spider
- Add support for cookies and authentication
- Implement robots.txt compliance
- Add more storage backends (e.g., database)
- Add support for custom JavaScript execution in headless mode
- Implement more sophisticated detection of websites requiring headless browsing
- Add automated captcha solving options (OCR, captcha solving services)
- Improve language selection for international websites
- Enhance RabbitMQ integration with more task types and better error handling

## License

MIT
