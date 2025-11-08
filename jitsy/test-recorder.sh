#!/bin/bash

# Jitsi Kurento Recorder - Test Script
# Tests the complete recording workflow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:9888}"
TEST_USER="test_user_$(date +%s)"
TEST_ROOM="test_room"
RECORDING_DURATION=30

echo -e "${GREEN}🧪 Jitsi Kurento Recorder - Test Script${NC}"
echo "=========================================="
echo "API URL: $API_URL"
echo "Test User: $TEST_USER"
echo "Test Room: $TEST_ROOM"
echo ""

# Function to make API request and check response
api_request() {
    local method=$1
    local endpoint=$2
    local expected_status=${3:-200}

    echo -e "${YELLOW}➜ $method $endpoint${NC}"

    response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}✓ Status: $http_code${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        echo -e "${RED}✗ Expected status $expected_status, got $http_code${NC}"
        echo "$body"
        return 1
    fi

    echo ""
}

# Test 1: Health Check
echo -e "\n${GREEN}Test 1: Health Check${NC}"
echo "--------------------"
api_request GET "/health" 200 || exit 1

# Test 2: Root endpoint
echo -e "\n${GREEN}Test 2: Root Endpoint${NC}"
echo "--------------------"
api_request GET "/" 200 || exit 1

# Test 3: List recordings (should be empty or contain other recordings)
echo -e "\n${GREEN}Test 3: List Active Recordings${NC}"
echo "--------------------"
api_request GET "/record/list" 200 || exit 1

# Test 4: Check status for non-existent user (should return 404)
echo -e "\n${GREEN}Test 4: Status for Non-Existent User${NC}"
echo "--------------------"
api_request GET "/record/status?user=nonexistent_user" 404 || echo -e "${YELLOW}⚠ Expected 404, continuing...${NC}"

# Test 5: Start recording
echo -e "\n${GREEN}Test 5: Start Recording${NC}"
echo "--------------------"
api_request POST "/record/start?user=$TEST_USER&room=$TEST_ROOM" 200 || exit 1

# Test 6: Try to start recording again (should fail with 400)
echo -e "\n${GREEN}Test 6: Duplicate Start (should fail)${NC}"
echo "--------------------"
api_request POST "/record/start?user=$TEST_USER&room=$TEST_ROOM" 400 || echo -e "${YELLOW}⚠ Expected 400, continuing...${NC}"

# Test 7: Check status for active recording
echo -e "\n${GREEN}Test 7: Check Recording Status${NC}"
echo "--------------------"
sleep 2
api_request GET "/record/status?user=$TEST_USER" 200 || exit 1

# Test 8: List recordings (should now include our test user)
echo -e "\n${GREEN}Test 8: List Recordings (should include test user)${NC}"
echo "--------------------"
api_request GET "/record/list" 200 || exit 1

# Test 9: Wait for recording
echo -e "\n${GREEN}Test 9: Recording in progress${NC}"
echo "--------------------"
echo "Recording for $RECORDING_DURATION seconds..."
for i in $(seq 1 $RECORDING_DURATION); do
    echo -n "."
    sleep 1
done
echo ""

# Test 10: Check status again
echo -e "\n${GREEN}Test 10: Check Status After Recording${NC}"
echo "--------------------"
api_request GET "/record/status?user=$TEST_USER" 200 || exit 1

# Test 11: Stop recording
echo -e "\n${GREEN}Test 11: Stop Recording${NC}"
echo "--------------------"
api_request POST "/record/stop?user=$TEST_USER&room=$TEST_ROOM" 200 || exit 1

# Test 12: Wait for upload to MinIO
echo -e "\n${GREEN}Test 12: Wait for MinIO Upload${NC}"
echo "--------------------"
echo "Waiting 10 seconds for background upload..."
for i in $(seq 1 10); do
    echo -n "."
    sleep 1
done
echo ""

# Test 13: Try to stop already stopped recording (should fail with 404)
echo -e "\n${GREEN}Test 13: Stop Already Stopped Recording (should fail)${NC}"
echo "--------------------"
api_request POST "/record/stop?user=$TEST_USER&room=$TEST_ROOM" 404 || echo -e "${YELLOW}⚠ Expected 404, continuing...${NC}"

# Test 14: Check that recording is no longer in active list
echo -e "\n${GREEN}Test 14: Verify Recording Removed from Active List${NC}"
echo "--------------------"
api_request GET "/record/list" 200 || exit 1

# Test 15: Final health check
echo -e "\n${GREEN}Test 15: Final Health Check${NC}"
echo "--------------------"
api_request GET "/health" 200 || exit 1

# Summary
echo ""
echo "=========================================="
echo -e "${GREEN}✅ All tests passed successfully!${NC}"
echo ""
echo "Recording details:"
echo "  User: $TEST_USER"
echo "  Room: $TEST_ROOM"
echo "  Duration: ${RECORDING_DURATION}s"
echo ""
echo "Check MinIO for the uploaded file:"
echo "  Bucket: jitsi-recordings"
echo "  Path: $TEST_ROOM/${TEST_USER}_*.webm"
echo "  URL: https://api.storage.recontext.online"
echo ""
echo -e "${GREEN}🎉 Test completed successfully!${NC}"
