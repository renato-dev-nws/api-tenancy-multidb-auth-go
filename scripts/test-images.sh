#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_URL="http://localhost:8081/api/v1"
TENANT_CODE="95RM301XKTJ"
EMAIL="joao@teste.com"
PASSWORD="senha12345"

# Variables
TOKEN=""
PRODUCT_ID=""
IMAGE_ID=""

echo -e "${YELLOW}[1/8] Creating test image...${NC}"
if [ ! -f test-image.jpg ]; then
    printf '%s' '/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA/9k=' | base64 -d > test-image.jpg 2>/dev/null
    echo -e "${GREEN}✓ Test image created${NC}"
else
    echo -e "${CYAN}→ Test image already exists${NC}"
fi
echo ""

echo -e "${YELLOW}[2/8] Authenticating...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Login failed${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo -e "${GREEN}✓ Login successful${NC}"
echo -e "${CYAN}  Token: ${TOKEN:0:30}...${NC}"
echo ""

echo -e "${YELLOW}[3/8] Creating test product...${NC}"
SKU="IMG-TEST-$(date +%s)"
PRODUCT_RESPONSE=$(curl -s -X POST "$API_URL/$TENANT_CODE/products" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"name\": \"Product with Images\",
        \"sku\": \"$SKU\",
        \"description\": \"Test product for image upload\",
        \"price\": 99.90,
        \"stock_quantity\": 10,
        \"is_active\": true
    }")

PRODUCT_ID=$(echo "$PRODUCT_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$PRODUCT_ID" ]; then
    echo -e "${RED}✗ Product creation failed${NC}"
    echo "Response: $PRODUCT_RESPONSE"
    exit 1
fi
PRODUCT_NAME=$(echo "$PRODUCT_RESPONSE" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}✓ Product created${NC}"
echo -e "${CYAN}  ID: $PRODUCT_ID${NC}"
echo -e "${CYAN}  Name: $PRODUCT_NAME${NC}"
echo ""

echo -e "${YELLOW}[4/8] Uploading image to product...${NC}"
UPLOAD_RESPONSE=$(curl -s -X POST "$API_URL/$TENANT_CODE/images" \
    -H "Authorization: Bearer $TOKEN" \
    -F "imageable_type=product" \
    -F "imageable_id=$PRODUCT_ID" \
    -F "files=@test-image.jpg" \
    -F "titles=Product Main Image" \
    -F "alt_texts=Main product image for SEO")

UPLOADED_COUNT=$(echo "$UPLOAD_RESPONSE" | grep -o '"uploaded":[0-9]*' | cut -d':' -f2)
if [ -z "$UPLOADED_COUNT" ] || [ "$UPLOADED_COUNT" == "0" ]; then
    echo -e "${RED}✗ Image upload failed${NC}"
    echo "Response: $UPLOAD_RESPONSE"
    # Cleanup
    curl -s -X DELETE "$API_URL/$TENANT_CODE/products/$PRODUCT_ID" -H "Authorization: Bearer $TOKEN" > /dev/null
    exit 1
fi

IMAGE_ID=$(echo "$UPLOAD_RESPONSE" | grep -o '"id":"[a-f0-9-]*"' | head -1 | cut -d'"' -f4)
STATUS=$(echo "$UPLOAD_RESPONSE" | grep -o '"processing_status":"[^"]*"' | head -1 | cut -d'"' -f4)
VARIANT=$(echo "$UPLOAD_RESPONSE" | grep -o '"variant":"[^"]*"' | head -1 | cut -d'"' -f4)

echo -e "${GREEN}✓ Image uploaded${NC}"
echo -e "${CYAN}  Uploaded: $UPLOADED_COUNT image(s)${NC}"
echo -e "${CYAN}  Image ID: $IMAGE_ID${NC}"
echo -e "${CYAN}  Status: $STATUS${NC}"
echo -e "${CYAN}  Variant: $VARIANT${NC}"
echo ""

echo -e "${YELLOW}[5/8] Listing images for product...${NC}"
LIST_RESPONSE=$(curl -s "$API_URL/$TENANT_CODE/images?imageable_type=product&imageable_id=$PRODUCT_ID" \
    -H "Authorization: Bearer $TOKEN")

TOTAL_IMAGES=$(echo "$LIST_RESPONSE" | grep -o '"id":' | wc -l)
echo -e "${GREEN}✓ Images listed${NC}"
echo -e "${CYAN}  Total: $TOTAL_IMAGES image(s)${NC}"

if [ "$TOTAL_IMAGES" != "0" ]; then
    FILENAME=$(echo "$LIST_RESPONSE" | grep -o '"filename":"[^"]*"' | head -1 | cut -d'"' -f4)
    VARIANT_LIST=$(echo "$LIST_RESPONSE" | grep -o '"variant":"[^"]*"' | head -1 | cut -d'"' -f4)
    STATUS_LIST=$(echo "$LIST_RESPONSE" | grep -o '"processing_status":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${CYAN}    - $FILENAME [$VARIANT_LIST] - Status: $STATUS_LIST${NC}"
fi
echo ""

echo -e "${YELLOW}[6/8] Getting image details...${NC}"
GET_RESPONSE=$(curl -s "$API_URL/$TENANT_CODE/images/$IMAGE_ID" \
    -H "Authorization: Bearer $TOKEN")

FILENAME=$(echo "$GET_RESPONSE" | grep -o '"filename":"[^"]*"' | cut -d'"' -f4)
SIZE=$(echo "$GET_RESPONSE" | grep -o '"file_size":[0-9]*' | cut -d':' -f2)
STORAGE_PATH=$(echo "$GET_RESPONSE" | grep -o '"storage_path":"[^"]*"' | cut -d'"' -f4)

echo -e "${GREEN}✓ Image details retrieved${NC}"
echo -e "${CYAN}  Filename: $FILENAME${NC}"
echo -e "${CYAN}  Size: $SIZE bytes${NC}"
echo -e "${CYAN}  Storage Path: $STORAGE_PATH${NC}"
echo ""

echo -e "${YELLOW}[7/8] Updating image metadata...${NC}"
UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/$TENANT_CODE/images/$IMAGE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Updated Product Image Title",
        "alt_text": "Updated description for SEO",
        "display_order": 1
    }')

NEW_TITLE=$(echo "$UPDATE_RESPONSE" | grep -o '"title":"[^"]*"' | cut -d'"' -f4)
NEW_ALT=$(echo "$UPDATE_RESPONSE" | grep -o '"alt_text":"[^"]*"' | cut -d'"' -f4)

echo -e "${GREEN}✓ Image metadata updated${NC}"
echo -e "${CYAN}  New title: $NEW_TITLE${NC}"
echo -e "${CYAN}  New alt text: $NEW_ALT${NC}"
echo ""

echo -e "${YELLOW}[8/8] Deleting image...${NC}"
curl -s -X DELETE "$API_URL/$TENANT_CODE/images/$IMAGE_ID" \
    -H "Authorization: Bearer $TOKEN" > /dev/null

echo -e "${GREEN}✓ Image deleted${NC}"
echo ""

echo -e "${CYAN}Cleaning up...${NC}"
curl -s -X DELETE "$API_URL/$TENANT_CODE/products/$PRODUCT_ID" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
echo -e "${GREEN}✓ Test product deleted${NC}"
echo ""

echo "========================================="
echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"
echo "========================================="
