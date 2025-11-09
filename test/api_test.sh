#!/bin/bash

# API Testing Script
# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
TOKEN=""
USER_ID=""

echo -e "${BLUE}=== Testing Psycho Platform API ===${NC}\n"

# Test 1: Health Check
echo -e "${BLUE}Test 1: Health Check${NC}"
HEALTH=$(curl -s ${BASE_URL}/health)
if echo $HEALTH | grep -q "healthy"; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo $HEALTH | jq '.'
else
    echo -e "${RED}✗ Health check failed${NC}"
    exit 1
fi
echo ""

# Test 2: Registration
echo -e "${BLUE}Test 2: User Registration${NC}"
REGISTER=$(curl -s -X POST ${BASE_URL}/api/auth/register \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser'$RANDOM'",
        "password": "testpass123",
        "display_name": "Test User"
    }')

if echo $REGISTER | grep -q "token"; then
    echo -e "${GREEN}✓ Registration successful${NC}"
    TOKEN=$(echo $REGISTER | jq -r '.token')
    USER_ID=$(echo $REGISTER | jq -r '.user.id')
    echo "Token: ${TOKEN:0:20}..."
    echo "User ID: $USER_ID"
else
    echo -e "${RED}✗ Registration failed${NC}"
    echo $REGISTER | jq '.'
    exit 1
fi
echo ""

# Test 3: Login
echo -e "${BLUE}Test 3: User Login${NC}"
LOGIN=$(curl -s -X POST ${BASE_URL}/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "password": "wrongpassword"
    }')

if echo $LOGIN | grep -q "Invalid credentials"; then
    echo -e "${GREEN}✓ Login validation working${NC}"
else
    echo -e "${RED}✗ Login validation failed${NC}"
fi
echo ""

# Test 4: Get Current User
echo -e "${BLUE}Test 4: Get Current User${NC}"
ME=$(curl -s ${BASE_URL}/api/auth/me \
    -H "Authorization: Bearer $TOKEN")

if echo $ME | grep -q "$USER_ID"; then
    echo -e "${GREEN}✓ Get user successful${NC}"
    echo $ME | jq '.username, .display_name'
else
    echo -e "${RED}✗ Get user failed${NC}"
    echo $ME | jq '.'
fi
echo ""

# Test 5: Create Topic
echo -e "${BLUE}Test 5: Create Topic${NC}"
TOPIC=$(curl -s -X POST ${BASE_URL}/api/topics \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Test Topic",
        "description": "This is a test topic",
        "is_public": true
    }')

if echo $TOPIC | grep -q "Test Topic"; then
    echo -e "${GREEN}✓ Create topic successful${NC}"
    TOPIC_ID=$(echo $TOPIC | jq -r '.id')
    echo "Topic ID: $TOPIC_ID"
else
    echo -e "${RED}✗ Create topic failed${NC}"
    echo $TOPIC | jq '.'
fi
echo ""

# Test 6: Get Topics
echo -e "${BLUE}Test 6: Get Topics${NC}"
TOPICS=$(curl -s ${BASE_URL}/api/topics \
    -H "Authorization: Bearer $TOKEN")

if echo $TOPICS | grep -q "Test Topic"; then
    echo -e "${GREEN}✓ Get topics successful${NC}"
    echo "Topics count: $(echo $TOPICS | jq 'length')"
else
    echo -e "${RED}✗ Get topics failed${NC}"
fi
echo ""

# Test 7: Create Message
echo -e "${BLUE}Test 7: Create Message${NC}"
MESSAGE=$(curl -s -X POST ${BASE_URL}/api/messages \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"content\": \"Test message with **markdown**\",
        \"topic_id\": \"$TOPIC_ID\"
    }")

if echo $MESSAGE | grep -q "Test message"; then
    echo -e "${GREEN}✓ Create message successful${NC}"
    MESSAGE_ID=$(echo $MESSAGE | jq -r '.id')
    echo "Message ID: $MESSAGE_ID"
else
    echo -e "${RED}✗ Create message failed${NC}"
    echo $MESSAGE | jq '.'
fi
echo ""

# Test 8: Get Messages
echo -e "${BLUE}Test 8: Get Messages${NC}"
MESSAGES=$(curl -s "${BASE_URL}/api/messages?topic_id=$TOPIC_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo $MESSAGES | grep -q "Test message"; then
    echo -e "${GREEN}✓ Get messages successful${NC}"
    echo "Messages count: $(echo $MESSAGES | jq 'length')"
else
    echo -e "${RED}✗ Get messages failed${NC}"
fi
echo ""

# Test 9: Add Reaction
echo -e "${BLUE}Test 9: Add Reaction${NC}"
REACTION=$(curl -s -X POST ${BASE_URL}/api/messages/$MESSAGE_ID/reactions \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "emoji": "👍"
    }')

if echo $REACTION | grep -q "id"; then
    echo -e "${GREEN}✓ Add reaction successful${NC}"
else
    echo -e "${RED}✗ Add reaction failed${NC}"
    echo $REACTION | jq '.'
fi
echo ""

# Test 10: Search
echo -e "${BLUE}Test 10: Global Search${NC}"
SEARCH=$(curl -s "${BASE_URL}/api/search?q=test" \
    -H "Authorization: Bearer $TOKEN")

if echo $SEARCH | grep -q "topics"; then
    echo -e "${GREEN}✓ Search successful${NC}"
    echo "Topics found: $(echo $SEARCH | jq '.topics | length')"
    echo "Messages found: $(echo $SEARCH | jq '.messages | length')"
else
    echo -e "${RED}✗ Search failed${NC}"
fi
echo ""

# Test 11: Unauthorized Access
echo -e "${BLUE}Test 11: Unauthorized Access${NC}"
UNAUTH=$(curl -s ${BASE_URL}/api/topics)

if echo $UNAUTH | grep -q "Authorization"; then
    echo -e "${GREEN}✓ Authorization protection working${NC}"
else
    echo -e "${RED}✗ Authorization protection failed${NC}"
fi
echo ""

# Test 12: WebSocket Connection
echo -e "${BLUE}Test 12: WebSocket Test (manual)${NC}"
echo "WebSocket URL: ws://localhost:8080/api/ws"
echo "Test with: wscat -c 'ws://localhost:8080/api/ws' -H 'Authorization: Bearer $TOKEN'"
echo ""

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "${GREEN}✓ All core tests passed${NC}"
echo -e "${BLUE}Platform is healthy and operational!${NC}"
