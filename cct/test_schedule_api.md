# Test Schedule API

## Prerequisites

1. Make sure the application is running:
```bash
make run
```

2. Make sure you have a novel in the database. If not, create one first:
```bash
# Create a website first
curl -X POST http://localhost:8080/api/websites \
  -H "Content-Type: application/json" \
  -d '{
    "name": "sangtacviet",
    "base_url": "https://sangtacviet.app",
    "script_name": "sangtacviet",
    "crawl_interval": 86400,
    "enabled": true
  }'

# Create a novel
curl -X POST http://localhost:8080/api/novels \
  -H "Content-Type: application/json" \
  -d '{
    "website_id": 1,
    "title": "Test Novel",
    "external_id": "test-novel-123",
    "source_url": "https://sangtacviet.app/truyen/test-novel-123"
  }'
```

## Test Schedule API

### 1. Create a Schedule

```bash
curl -X POST http://localhost:8080/api/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "novel_id": 1,
    "enabled": true,
    "interval_seconds": 3600
  }'
```

Expected response:
```json
{
  "id": 1,
  "novel_id": 1,
  "enabled": true,
  "interval_seconds": 3600,
  "last_run_at": null,
  "next_run_at": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T11:00:00Z",
  "updated_at": "2024-01-01T11:00:00Z"
}
```

### 2. Get All Schedules

```bash
curl http://localhost:8080/api/schedules
```

### 3. Get Schedules for a Novel

```bash
curl http://localhost:8080/api/schedules?novel_id=1
```

### 4. Get a Specific Schedule

```bash
curl http://localhost:8080/api/schedules/1
```

### 5. Update a Schedule

```bash
curl -X PUT http://localhost:8080/api/schedules/1 \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": false,
    "interval_seconds": 7200
  }'
```

### 6. Trigger a Schedule Immediately

```bash
curl -X POST http://localhost:8080/api/schedules/1/trigger
```

Expected response:
```json
{
  "status": "success",
  "message": "Schedule triggered successfully",
  "schedule": {
    "id": 1,
    "novel_id": 1,
    "enabled": true,
    "interval_seconds": 3600,
    "last_run_at": null,
    "next_run_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T11:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### 7. Delete a Schedule

```bash
curl -X DELETE http://localhost:8080/api/schedules/1
```

## Test Scheduler Functionality

### 1. Create a Schedule with Long Interval

```bash
curl -X POST http://localhost:8080/api/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "novel_id": 1,
    "enabled": true,
    "interval_seconds": 3600
  }'
```

### 2. Trigger the Schedule Immediately

Instead of waiting for the schedule to run naturally, trigger it immediately:

```bash
curl -X POST http://localhost:8080/api/schedules/1/trigger
```

This will update the `next_run_at` to current time, so the scheduler will pick it up within 60 seconds.

### 3. Monitor Logs

Watch the application logs to see the scheduler in action:
- It should check for due schedules every 60 seconds (configurable)
- When a schedule is due, it should publish a book crawl task
- After book crawl completes, it should create chapter crawl tasks

### 4. Check Chapter Crawl Logs

After some crawling activity:
```bash
curl http://localhost:8080/api/chapters/1/logs
```

## Expected Behavior

1. **Schedule Creation**: Creates a record in `novel_schedules` table
2. **Scheduler**: Runs every 60 seconds and checks for due schedules
3. **Book Crawl**: When due, publishes book crawl task to RabbitMQ
4. **Chapter Crawl**: After book crawl, creates chapter crawl tasks for empty chapters
5. **Logging**: Logs all chapter crawl results to `chapter_crawl_logs` table
6. **Time Updates**: Updates `last_run_at` and `next_run_at` after each run

## Troubleshooting

1. **No schedules running**: Check if scheduler is enabled in config.yml
2. **No agents**: Make sure you have active crawler agents
3. **Database errors**: Check if migrations were applied correctly
4. **RabbitMQ errors**: Make sure RabbitMQ is running and accessible
