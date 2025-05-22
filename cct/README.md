# Crawler Control API

This is a REST API for controlling web crawlers, managing websites to crawl, novels, chapters, and crawl jobs.

## Setup

### Prerequisites

- Go 1.22 or higher
- PostgreSQL database

### Configuration

The application is configured using a `config.yml` file. Here's an example configuration:

```yaml
# Server configuration
server:
  port: 8080
  read_timeout: 15  # seconds
  write_timeout: 15 # seconds
  idle_timeout: 60  # seconds

# Database configuration
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: crawler
  sslmode: disable

# Authentication configuration
auth:
  enabled: false
  token_expiry: 720 # hours (30 days)

# Logging configuration
logging:
  level: info      # debug, info, warn, error
  format: text     # text or json
  output: console  # console, file, or both
  file_path: logs/app.log
  max_size: 10     # maximum size in MB before rotation
  max_backups: 5   # maximum number of old log files to retain
  max_age: 30      # maximum number of days to retain old log files
  compress: true   # compress rotated files

# RabbitMQ configuration
rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
  queue_name: "crawler_tasks"
  exchange_name: "crawler_exchange"
  exchange_type: "topic"
  routing_keys:
    - "crawl.sangtacviet.book"
    - "crawl.sangtacviet.chapter"
    - "crawl.sangtacviet.session"
    - "crawl.wikidich.book"
    - "crawl.wikidich.chapter"
    - "crawl.metruyenchu.book"
    - "crawl.metruyenchu.chapter"
  prefetch_count: 1
  reconnect_interval: 5 # seconds
```

You can also override these settings using environment variables. For example, to change the database host, you can set the `DATABASE_HOST` environment variable.

### Database Setup

1. Create a PostgreSQL database
2. Run the SQL schema using the Makefile:

```bash
# Set the database name
export DB_NAME=crawler

# Create the schema
make db-schema
```

Alternatively, you can run the SQL schema manually:

```bash
psql -d crawler -f crawler_schema.sql
```

### Running the API

You can use the provided Makefile to build and run the application:

```bash
# Install dependencies
make deps

# Build the application
make build

# Run the application
make run

# Or simply build and run in one command
make && ./build/cct
```

For more Makefile commands, run:

```bash
make help
```

## API Endpoints

### Websites

- `GET /api/websites`: Get all websites
- `GET /api/websites/{id}`: Get a website by ID
- `POST /api/websites`: Create a new website
- `PUT /api/websites/{id}`: Update a website
- `DELETE /api/websites/{id}`: Delete a website

### Novels

- `GET /api/novels`: Get all novels
- `GET /api/novels?website_id={id}`: Get novels for a specific website
- `GET /api/novels/{id}`: Get a novel by ID
- `POST /api/novels`: Create a new novel
- `PUT /api/novels/{id}`: Update a novel
- `DELETE /api/novels/{id}`: Delete a novel

### Chapters

- `GET /api/chapters`: Get all chapters
- `GET /api/chapters?novel_id={id}`: Get chapters for a specific novel
- `GET /api/chapters/{id}`: Get a chapter by ID
- `POST /api/chapters`: Create a new chapter
- `PUT /api/chapters/{id}`: Update a chapter
- `DELETE /api/chapters/{id}`: Delete a chapter

### Crawl Jobs

- `GET /api/crawl-jobs`: Get all crawl jobs
- `GET /api/crawl-jobs?novel_id={id}`: Get crawl jobs for a specific novel
- `GET /api/crawl-jobs/{id}`: Get a crawl job by ID
- `POST /api/crawl-jobs`: Create a new crawl job
- `PUT /api/crawl-jobs/{id}`: Update a crawl job
- `DELETE /api/crawl-jobs/{id}`: Delete a crawl job
- `POST /api/crawl-jobs/{id}/start`: Start a crawl job
- `POST /api/crawl-jobs/{id}/complete`: Mark a crawl job as completed
- `POST /api/crawl-jobs/{id}/fail`: Mark a crawl job as failed with an error message

### Agents

- `GET /api/agents`: Get all agents
- `GET /api/agents?active_only=true`: Get only active agents
- `GET /api/agents/{id}`: Get an agent by ID
- `POST /api/agents`: Create a new agent
- `PUT /api/agents/{id}`: Update an agent
- `DELETE /api/agents/{id}`: Delete an agent
- `POST /api/agents/{id}/heartbeat`: Update agent heartbeat
- `POST /api/agents/deactivate-inactive`: Deactivate inactive agents

### Users

- `GET /api/users`: Get all users
- `GET /api/users/{id}`: Get a user by ID
- `POST /api/users`: Create a new user
- `PUT /api/users/{id}`: Update a user
- `PUT /api/users/{id}/password`: Update a user's password
- `DELETE /api/users/{id}`: Delete a user

### API Tokens

- `GET /api/api-tokens`: Get all API tokens
- `GET /api/api-tokens?user_id={id}`: Get API tokens for a specific user
- `GET /api/api-tokens/{id}`: Get an API token by ID
- `POST /api/api-tokens`: Create a new API token
- `PUT /api/api-tokens/{id}`: Update an API token
- `DELETE /api/api-tokens/{id}`: Delete an API token
- `POST /api/api-tokens/delete-expired`: Delete all expired API tokens

### RabbitMQ Tasks

- `POST /api/tasks/publish`: Publish a task to active agents
- `GET /api/agents/count`: Get the count of active agents

### Authentication

- `POST /api/auth/login`: Login with email and password to get an API token
- `POST /api/auth/register`: Register a new user account

## RabbitMQ Integration

The application includes RabbitMQ integration for distributing tasks to crawler agents. Only active agents will receive messages.

### Publishing Tasks

To publish a task to active agents, use the publish endpoint:

```
POST /api/tasks/publish
```

Request body:
```json
{
  "source": "sangtacviet",
  "task_type": "book",
  "url": "https://sangtacviet.app/truyen/12345",
  "timeout_sec": 30
}
```

Parameters:
- `source`: The source of the task (sangtacviet, wikidich, metruyenchu)
- `task_type`: The type of task (book, chapter, session)
- `url`: The URL to crawl
- `timeout_sec`: (Optional) Timeout in seconds for the task (default: 30)

Response:
```json
{
  "status": "success",
  "message": "Task published successfully"
}
```

### Getting Active Agent Count

To get the count of active agents, use the count endpoint:

```
GET /api/agents/count
```

Response:
```json
{
  "count": 5
}
```

## Authentication

Authentication is controlled by the `auth.enabled` setting in the configuration file. When enabled, all API requests (except for the login and register endpoints) must include an `Authorization` header with a valid API token.

### Registration

To create a new user account, use the register endpoint:

```
POST /api/auth/register
```

Request body:
```json
{
  "email": "user@example.com",
  "password": "your_password"
}
```

Response:
```json
{
  "id": 1,
  "email": "user@example.com",
  "created_at": "2023-06-01T00:00:00Z"
}
```

### Login

To obtain an API token, use the login endpoint:

```
POST /api/auth/login
```

Request body:
```json
{
  "email": "user@example.com",
  "password": "your_password"
}
```

Response:
```json
{
  "token": "your_api_token",
  "expires_at": "2023-06-01T00:00:00Z",
  "user_id": 1
}
```

### Using the API Token

Include the token in the `Authorization` header for all subsequent requests:

```
Authorization: Bearer your_api_token
```

API tokens can be managed through the API endpoints. They expire after 30 days by default, but this can be configured in the `auth.token_expiry` setting.

## Logging

The application uses a structured logger that can write to both the console and files. The logging configuration is controlled by the `logging` section in the configuration file.

### Log Levels

The following log levels are supported, in order of increasing severity:

- `debug`: Detailed information, typically of interest only when diagnosing problems
- `info`: Confirmation that things are working as expected
- `warn`: Indication that something unexpected happened, but the application can continue
- `error`: Due to a more serious problem, the application has not been able to perform a function
- `fatal`: Very severe error events that will lead the application to abort

### Log Outputs

The logger can write to the following outputs:

- `console`: Logs are written to the console in a human-readable format
- `file`: Logs are written to a file with automatic rotation
- `both`: Logs are written to both the console and a file

### Log Rotation

When logging to a file, the following rotation settings are available:

- `max_size`: Maximum size of the log file in megabytes before it gets rotated
- `max_backups`: Maximum number of old log files to retain
- `max_age`: Maximum number of days to retain old log files
- `compress`: Whether to compress rotated files
