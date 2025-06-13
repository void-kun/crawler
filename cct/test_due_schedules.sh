#!/bin/bash

# Test script to debug GetDueSchedules function
BASE_URL="http://localhost:8080"

echo "=== Testing GetDueSchedules Function ==="
echo

# Function to make HTTP requests
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
    
    if [[ $status_code -ge 200 && $status_code -lt 300 ]]; then
        echo "✅ Success!"
    else
        echo "❌ Failed!"
    fi
    echo "----------------------------------------"
    echo
}

echo "Step 1: Create a website (if not exists)..."
website_data='{
    "name": "sangtacviet",
    "base_url": "https://sangtacviet.app",
    "script_name": "sangtacviet",
    "crawl_interval": 86400,
    "enabled": true
}'
make_request "POST" "$BASE_URL/api/websites" "$website_data" "Creating website"

echo "Step 2: Create a novel (if not exists)..."
novel_data='{
    "website_id": 1,
    "title": "Test Novel for Due Schedule",
    "external_id": "test-novel-due-123",
    "source_url": "https://sangtacviet.app/truyen/test-novel-due-123"
}'
make_request "POST" "$BASE_URL/api/novels" "$novel_data" "Creating novel"

echo "Step 3: Create a schedule with short interval..."
schedule_data='{
    "novel_id": 1,
    "enabled": true,
    "interval_seconds": 60
}'
make_request "POST" "$BASE_URL/api/schedules" "$schedule_data" "Creating schedule"

echo "Step 4: Get all schedules to see current state..."
make_request "GET" "$BASE_URL/api/schedules" "" "Getting all schedules"

echo "Step 5: Trigger the schedule to make it due..."
make_request "POST" "$BASE_URL/api/schedules/1/trigger" "" "Triggering schedule"

echo "Step 6: Get schedules again to see triggered state..."
make_request "GET" "$BASE_URL/api/schedules" "" "Getting schedules after trigger"

echo "Step 7: Test GetDueSchedules function directly..."
make_request "GET" "$BASE_URL/api/schedules/due" "" "Getting due schedules"

echo "=== Manual Database Check ==="
echo
echo "You can manually check the database with these queries:"
echo
echo "1. Check all schedules:"
echo "   SELECT id, novel_id, enabled, next_run_at, NOW() as current_time"
echo "   FROM novel_schedules;"
echo
echo "2. Check due schedules (same query as GetDueSchedules):"
echo "   SELECT id, novel_id, enabled, interval_seconds, last_run_at, next_run_at, created_at, updated_at"
echo "   FROM novel_schedules"
echo "   WHERE enabled = true AND next_run_at IS NOT NULL AND next_run_at <= NOW()"
echo "   ORDER BY next_run_at;"
echo
echo "3. Check with comparison:"
echo "   SELECT id, enabled, next_run_at, NOW() as current_time,"
echo "          CASE WHEN next_run_at <= NOW() THEN 'DUE' ELSE 'NOT_DUE' END as status"
echo "   FROM novel_schedules;"
echo
echo "=== Expected Behavior ==="
echo "- After triggering, next_run_at should be set to current time"
echo "- The schedule should appear in GetDueSchedules results"
echo "- Scheduler should pick it up within 60 seconds"
