#!/bin/bash

# Demo script for testing schedule trigger API
# Make sure the application is running before executing this script

BASE_URL="http://localhost:8080"

echo "=== Schedule Trigger API Demo ==="
echo

# Function to make HTTP requests with error handling
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4
    
    echo "📡 $description"
    echo "   $method $url"
    
    if [ -n "$data" ]; then
        echo "   Data: $data"
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    fi
    
    # Extract response body and status code
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    echo "   Status: $status_code"
    echo "   Response: $body"
    echo
    
    # Check if request was successful
    if [[ $status_code -ge 200 && $status_code -lt 300 ]]; then
        echo "✅ Success!"
    else
        echo "❌ Failed!"
    fi
    echo "----------------------------------------"
    echo
}

# Step 1: Create a website (if not exists)
echo "Step 1: Creating a website..."
website_data='{
    "name": "sangtacviet",
    "base_url": "https://sangtacviet.app",
    "script_name": "sangtacviet",
    "crawl_interval": 86400,
    "enabled": true
}'
make_request "POST" "$BASE_URL/api/websites" "$website_data" "Creating website"

# Step 2: Create a novel (if not exists)
echo "Step 2: Creating a novel..."
novel_data='{
    "website_id": 1,
    "title": "Test Novel for Schedule",
    "external_id": "test-novel-schedule-123",
    "source_url": "https://sangtacviet.app/truyen/test-novel-schedule-123"
}'
make_request "POST" "$BASE_URL/api/novels" "$novel_data" "Creating novel"

# Step 3: Create a schedule with long interval
echo "Step 3: Creating a schedule with long interval (1 hour)..."
schedule_data='{
    "novel_id": 1,
    "enabled": true,
    "interval_seconds": 3600
}'
make_request "POST" "$BASE_URL/api/schedules" "$schedule_data" "Creating schedule"

# Step 4: Get the created schedule
echo "Step 4: Getting the created schedule..."
make_request "GET" "$BASE_URL/api/schedules/1" "" "Getting schedule details"

# Step 5: Trigger the schedule immediately
echo "Step 5: Triggering the schedule immediately..."
make_request "POST" "$BASE_URL/api/schedules/1/trigger" "" "Triggering schedule"

# Step 6: Get the schedule again to see updated times
echo "Step 6: Getting schedule after trigger..."
make_request "GET" "$BASE_URL/api/schedules/1" "" "Getting updated schedule"

# Step 7: List all schedules
echo "Step 7: Listing all schedules..."
make_request "GET" "$BASE_URL/api/schedules" "" "Listing all schedules"

echo "=== Demo Complete ==="
echo
echo "📋 What happened:"
echo "1. Created a website and novel (if they didn't exist)"
echo "2. Created a schedule with 1-hour interval"
echo "3. Triggered the schedule to run immediately"
echo "4. The next_run_at time was updated to current time"
echo "5. The scheduler will pick this up within 60 seconds"
echo
echo "🔍 To monitor the scheduler:"
echo "- Watch the application logs for scheduler activity"
echo "- Check if book crawl tasks are published to RabbitMQ"
echo "- Verify that last_run_at and next_run_at are updated after execution"
echo
echo "🧪 Additional tests you can run:"
echo "- curl $BASE_URL/api/schedules/1 (check schedule status)"
echo "- curl $BASE_URL/api/chapters/1/logs (check crawl logs after execution)"
echo "- curl -X POST $BASE_URL/api/schedules/1/trigger (trigger again)"
