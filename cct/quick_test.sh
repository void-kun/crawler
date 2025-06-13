#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== Quick Test GetDueSchedules ==="
echo

echo "1. Testing GetDueSchedules endpoint..."
curl -s "$BASE_URL/api/schedules/due" | jq '.'

echo
echo "2. Testing all schedules..."
curl -s "$BASE_URL/api/schedules" | jq '.'

echo
echo "3. Trigger schedule 1 (if exists)..."
curl -s -X POST "$BASE_URL/api/schedules/1/trigger" | jq '.'

echo
echo "4. Test GetDueSchedules again..."
curl -s "$BASE_URL/api/schedules/due" | jq '.'

echo
echo "=== Check application logs for DEBUG output ==="
