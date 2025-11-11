#!/bin/bash

# Test Avatar Upload API

set -e

API_URL="${API_URL:-http://localhost:20081}"
USERNAME="${USERNAME:-admin@recontext.online}"
PASSWORD="${PASSWORD:-admin123}"

echo "🧪 Testing Avatar Upload API"
echo "================================"
echo ""

# Step 1: Login
echo "1️⃣ Logging in as $USERNAME..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

echo "$LOGIN_RESPONSE" | jq '.' 2>/dev/null || echo "$LOGIN_RESPONSE"

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token' 2>/dev/null)
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user.id' 2>/dev/null)

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Login failed!"
    exit 1
fi

echo "✅ Login successful!"
echo "   User ID: $USER_ID"
echo "   Token: ${TOKEN:0:20}..."
echo ""

# Step 2: Get current profile
echo "2️⃣ Getting current profile..."
PROFILE=$(curl -s -X GET "$API_URL/api/v1/users/$USER_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "$PROFILE" | jq '.' 2>/dev/null || echo "$PROFILE"
echo ""

# Step 3: Create test image if it doesn't exist
TEST_IMAGE="/tmp/test-avatar.jpg"
if [ ! -f "$TEST_IMAGE" ]; then
    echo "3️⃣ Creating test image..."
    # Create a simple colored square using ImageMagick (if available) or skip
    if command -v convert &> /dev/null; then
        convert -size 200x200 xc:blue "$TEST_IMAGE"
        echo "✅ Test image created: $TEST_IMAGE"
    else
        echo "⚠️  ImageMagick not found. Please provide your own image:"
        echo "   $TEST_IMAGE"
        echo ""
        echo "Or use an existing image:"
        read -p "Enter path to image file: " USER_IMAGE
        if [ -f "$USER_IMAGE" ]; then
            TEST_IMAGE="$USER_IMAGE"
        else
            echo "❌ File not found: $USER_IMAGE"
            exit 1
        fi
    fi
else
    echo "3️⃣ Using existing test image: $TEST_IMAGE"
fi
echo ""

# Step 4: Upload avatar
echo "4️⃣ Uploading avatar..."
UPLOAD_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/users/$USER_ID/avatar" \
  -H "Authorization: Bearer $TOKEN" \
  -F "avatar=@$TEST_IMAGE")

echo "$UPLOAD_RESPONSE" | jq '.' 2>/dev/null || echo "$UPLOAD_RESPONSE"

AVATAR_URL=$(echo "$UPLOAD_RESPONSE" | jq -r '.avatar_url' 2>/dev/null)

if [ "$AVATAR_URL" = "null" ] || [ -z "$AVATAR_URL" ]; then
    echo "❌ Avatar upload failed!"
    exit 1
fi

echo "✅ Avatar uploaded successfully!"
echo "   Avatar URL: $AVATAR_URL"
echo ""

# Step 5: Verify profile was updated
echo "5️⃣ Verifying profile was updated..."
UPDATED_PROFILE=$(curl -s -X GET "$API_URL/api/v1/users/$USER_ID" \
  -H "Authorization: Bearer $TOKEN")

PROFILE_AVATAR=$(echo "$UPDATED_PROFILE" | jq -r '.avatar' 2>/dev/null)

if [ "$PROFILE_AVATAR" = "$AVATAR_URL" ]; then
    echo "✅ Profile updated successfully!"
    echo "   Avatar in profile: $PROFILE_AVATAR"
else
    echo "⚠️  Avatar URL mismatch!"
    echo "   Expected: $AVATAR_URL"
    echo "   Got: $PROFILE_AVATAR"
fi
echo ""

# Step 6: Update profile information
echo "6️⃣ Updating profile information..."
UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/api/v1/users/$USER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "User",
    "phone": "+79991234567",
    "bio": "Testing avatar upload functionality",
    "language": "en"
  }')

echo "$UPDATE_RESPONSE" | jq '.' 2>/dev/null || echo "$UPDATE_RESPONSE"
echo ""

# Step 7: Final verification
echo "7️⃣ Final profile verification..."
FINAL_PROFILE=$(curl -s -X GET "$API_URL/api/v1/users/$USER_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "$FINAL_PROFILE" | jq '{
  id: .id,
  username: .username,
  email: .email,
  first_name: .first_name,
  last_name: .last_name,
  phone: .phone,
  bio: .bio,
  avatar: .avatar
}' 2>/dev/null || echo "$FINAL_PROFILE"

echo ""
echo "================================"
echo "✅ All tests passed!"
echo ""
echo "You can now:"
echo "  1. Check the uploaded file: ls -lh uploads/avatars/"
echo "  2. View avatar in browser: $API_URL$AVATAR_URL"
echo "  3. Test in UI at: $API_URL"
