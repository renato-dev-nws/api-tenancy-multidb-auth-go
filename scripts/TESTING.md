# Image Upload System - Testing Guide

## 🚀 Quick Start

### 1. Create Test Image
```bash
make create-test-image
```
Generates a minimal 1x1 pixel JPEG for testing.

### 2. Run Complete Workflow Test
```bash
make test-images-complete
```

This will:
- ✅ Create test image
- ✅ Login as tenant user
- ✅ Create a test product
- ✅ Upload image to product
- ✅ List uploaded images
- ✅ Retrieve image details
- ✅ Update image metadata
- ✅ Delete image
- ✅ Cleanup test data

**Expected output:**
```
=========================================
Image Upload - Complete Workflow Test
=========================================

[1/8] Creating test image...
✓ Test image created

[2/8] Authenticating...
✓ Login successful
  Token: eyJhbGciOiJIUzI1NiI...

[3/8] Creating test product...
✓ Product created
  ID: 123e4567-e89b-12d3-a456-426614174000
  Name: Product with Images

[4/8] Uploading image to product...
✓ Image uploaded
  Uploaded: 1 image(s)
  Image ID: 456e7890-e89b-12d3-a456-426614174000
  Status: pending
  Variant: original

[5/8] Listing images for product...
✓ Images listed
  Total: 1 image(s)
    - 456e7890_original.jpg [original] - Status: pending

[6/8] Getting image details...
✓ Image details retrieved
  Filename: 456e7890_original.jpg
  Size: 635 bytes
  Storage Path: tenant-uuid/images/product/123e4567/456e7890_original.jpg

[7/8] Updating image metadata...
✓ Image metadata updated
  New title: Updated Product Image Title
  New alt text: Updated description for SEO

[8/8] Deleting image...
✓ Image deleted

Cleaning up...
✓ Test product deleted

=========================================
✓ ALL TESTS PASSED
=========================================
```

## 📝 Individual Tests

### Upload Image to Existing Product

```bash
# First, get a product ID
make test-product-create

# Then upload image
make test-image-upload PRODUCT_ID=<uuid>
```

### List Product Images

```bash
make test-image-list PRODUCT_ID=<uuid>
```

### Get Image Details

```bash
make test-image-get IMAGE_ID=<uuid>
```

### Update Image Metadata

```bash
make test-image-update IMAGE_ID=<uuid>
```

### Delete Image

```bash
make test-image-delete IMAGE_ID=<uuid>
```

## 🔍 Verifying Worker Processing

### Check Worker Status

```bash
# Check if worker is running
docker compose ps image-worker

# View worker logs
make worker-logs
# or
docker compose logs -f image-worker
```

### Monitor Image Processing

After uploading an image, check its processing status:

```bash
# Get image details
make test-image-get IMAGE_ID=<uuid>
```

**Processing States:**
- `pending` - Waiting for worker to process
- `processing` - Worker is currently processing
- `completed` - All variants generated successfully
- `failed` - Processing error occurred

### Verify Variants Were Created

```bash
# List images - should show original + variants
make test-image-list PRODUCT_ID=<uuid>
```

Expected variants:
- `original` (1400x1400 max)
- `medium` (800x800 max)
- `small` (350x350 max)
- `thumb` (100x100 max)

### Check File System

```bash
# View uploaded files
ls -R uploads/

# Expected structure:
# uploads/
#   {tenant-uuid}/
#     images/
#       product/
#         {product-id}/
#           {image-id}_original.jpg
#           {image-id}_medium.jpg
#           {image-id}_small.jpg
#           {image-id}_thumb.jpg
```

## 🐛 Troubleshooting

### Test Image Not Found

```bash
make create-test-image
```

### Authentication Failed

Check if user exists:
```bash
make test-login
```

Expected credentials:
- Email: `joao@teste.com`
- Password: `senha12345`

### Product Not Found

Ensure product exists or create one:
```bash
make test-product-create
```

### Worker Not Processing

1. Check worker is running:
```bash
docker compose ps image-worker
```

2. Check Redis connection:
```bash
docker compose exec redis redis-cli PING
```

3. Check worker logs for errors:
```bash
docker compose logs image-worker --tail 50
```

4. Manually trigger processing (if worker crashed):
```sql
UPDATE images 
SET processing_status = 'pending' 
WHERE processing_status = 'processing';
```

### Permission Denied

The image routes use product/service permissions:
- Upload: Requires `manage_products` permission
- Update: Requires `manage_products` permission
- Delete: Requires `delete_product` permission
- List/Get: No permission required (read-only)

Ensure user has these permissions in their role.

### Storage Directory Not Writable

```bash
# Fix permissions
chmod -R 755 uploads/

# Or in Docker
docker compose exec tenant-api chmod -R 755 /app/uploads
```

## 📊 Performance Testing

### Upload Multiple Images

```bash
# Create 10 test images
for i in {1..10}; do
    make create-test-image
    mv test-image.jpg test-image-$i.jpg
done

# Upload all (requires custom script)
# See scripts/bulk-upload.ps1
```

### Benchmark Worker Processing

```bash
# Get current timestamp
$start = Get-Date

# Upload image
make test-image-upload PRODUCT_ID=<uuid>

# Wait for completion
while ($true) {
    $status = (Invoke-RestMethod -Uri "http://localhost:8081/api/v1/95RM301XKTJ/images/$imageId" -Headers @{Authorization="Bearer $token"}).processing_status
    if ($status -eq "completed") { break }
    Start-Sleep -Seconds 1
}

$end = Get-Date
$duration = ($end - $start).TotalSeconds
Write-Host "Processing time: $duration seconds"
```

Expected: ~1-2 seconds for small image

## 🔐 Security Testing

### Test Cross-Tenant Isolation

```bash
# Upload image to Tenant A
make test-image-upload PRODUCT_ID=<tenant-a-product>

# Try to access from Tenant B (should fail)
curl -X GET "http://localhost:8081/api/v1/<tenant-b-code>/images/<tenant-a-image-id>" \
  -H "Authorization: Bearer <tenant-b-token>"
```

Expected: 404 Not Found (image isolation working)

### Test Permission Enforcement

```bash
# Login as user without manage_products permission
# Try to upload (should fail with 403)
```

## 📚 Additional Resources

- **Worker Documentation**: [cmd/image-worker/README.md](../cmd/image-worker/README.md)
- **API Documentation**: Check Swagger/OpenAPI docs
- **Storage Configuration**: See `.env.example` for S3/R2 setup

## 🎯 CI/CD Integration

### GitHub Actions Example

```yaml
- name: Test Image Upload
  run: |
    make setup
    make create-test-image
    make test-images-complete
```

### Expected Test Duration

- Quick test: ~2 seconds
- Complete workflow: ~5-10 seconds
- With worker processing: +2-3 seconds per image

---

**Last updated**: February 2026  
**Version**: 1.0.0
