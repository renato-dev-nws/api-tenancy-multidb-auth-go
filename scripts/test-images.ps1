# Complete Image Upload Workflow Test
# Tests: Login → Create Product → Upload Image → List → Update → Delete

$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8081"
$apiPath = "/api/v1"
$tenantCode = "95RM301XKTJ"

# Colors
function Write-Success { param($msg) Write-Host $msg -ForegroundColor Green }
function Write-Info { param($msg) Write-Host $msg -ForegroundColor Cyan }
function Write-Error { param($msg) Write-Host $msg -ForegroundColor Red }
function Write-Step { param($step, $msg) Write-Host "`n[$step] $msg" -ForegroundColor Yellow }

Write-Host "`n=========================================" -ForegroundColor Cyan
Write-Host "Image Upload - Complete Workflow Test" -ForegroundColor Cyan
Write-Host "=========================================`n" -ForegroundColor Cyan

try {
    # Step 1: Create test image if not exists
    Write-Step "1/8" "Creating test image..."
    if (-not (Test-Path "test-image.jpg")) {
        $jpegBytes = [Convert]::FromBase64String('/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAA/9k=')
        [System.IO.File]::WriteAllBytes("test-image.jpg", $jpegBytes)
        Write-Success "✓ Test image created"
    } else {
        Write-Success "✓ Test image already exists"
    }

    # Step 2: Login
    Write-Step "2/8" "Authenticating..."
    $loginBody = @{
        email = "joao@teste.com"
        password = "senha12345"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$baseUrl$apiPath/auth/login" `
        -Method Post -ContentType "application/json" -Body $loginBody

    $token = $loginResponse.token
    $headers = @{
        Authorization = "Bearer $token"
    }
    Write-Success "✓ Login successful"
    Write-Info "  Token: $($token.Substring(0, 20))..."

    # Step 3: Create a test product
    Write-Step "3/8" "Creating test product..."
    $productBody = @{
        name = "Product with Images"
        description = "Test product for image upload workflow"
        price = 99.99
        sku = "IMG-TEST-$(Get-Random -Minimum 1000 -Maximum 9999)"
        stock = 10
    } | ConvertTo-Json

    $productResponse = Invoke-RestMethod -Uri "$baseUrl$apiPath/$tenantCode/products" `
        -Method Post -Headers $headers -ContentType "application/json" -Body $productBody

    $productId = $productResponse.id
    Write-Success "✓ Product created"
    Write-Info "  ID: $productId"
    Write-Info "  Name: $($productResponse.name)"

    # Step 4: Upload image
    Write-Step "4/8" "Uploading image to product..."
    $form = @{
        imageable_type = "product"
        imageable_id = $productId
        files = Get-Item "test-image.jpg"
        titles = "Test Product Image"
        alt_texts = "Beautiful product photo"
    }

    $uploadResponse = Invoke-RestMethod -Uri "$baseUrl$apiPath/$tenantCode/images" `
        -Method Post -Headers $headers -Form $form

    Write-Success "✓ Image uploaded"
    Write-Info "  Uploaded: $($uploadResponse.uploaded) image(s)"
    
    if ($uploadResponse.images -and $uploadResponse.images.Count -gt 0) {
        $imageId = $uploadResponse.images[0].id
        $imageStatus = $uploadResponse.images[0].processing_status
        Write-Info "  Image ID: $imageId"
        Write-Info "  Status: $imageStatus"
        Write-Info "  Variant: $($uploadResponse.images[0].variant)"
    } else {
        throw "No images returned in upload response"
    }

    # Step 5: List images for product
    Write-Step "5/8" "Listing images for product..."
    $listResponse = Invoke-RestMethod -Uri "$baseUrl$apiPath/$tenantCode/images?imageable_type=product&imageable_id=$productId" `
        -Headers $headers

    Write-Success "✓ Images listed"
    Write-Info "  Total: $($listResponse.Count) image(s)"
    if ($listResponse.Count -gt 0) {
        foreach ($img in $listResponse) {
            Write-Info "    - $($img.filename) [$($img.variant)] - Status: $($img.processing_status)"
        }
    }

    # Step 6: Get single image
    Write-Step "6/8" "Getting image details..."
    $imageResponse = Invoke-RestMethod -Uri "$baseUrl$apiPath/$tenantCode/images/$imageId" `
        -Headers $headers

    Write-Success "✓ Image details retrieved"
    Write-Info "  Filename: $($imageResponse.filename)"
    Write-Info "  Size: $($imageResponse.file_size) bytes"
    Write-Info "  Storage Path: $($imageResponse.storage_path)"
    
    # Check for variants
    if ($imageResponse.variants) {
        Write-Info "  Variants: $($imageResponse.variants.Count)"
    }

    # Step 7: Update image metadata
    Write-Step "7/8" "Updating image metadata..."
    $updateBody = @{
        title = "Updated Product Image Title"
        alt_text = "Updated description for SEO"
        display_order = 1
    } | ConvertTo-Json

    $updateResponse = Invoke-RestMethod -Uri "$baseUrl$apiPath/$tenantCode/images/$imageId" `
        -Method Put -Headers ($headers + @{"Content-Type" = "application/json"}) -Body $updateBody

    Write-Success "✓ Image metadata updated"
    Write-Info "  New title: $($updateResponse.title)"
    Write-Info "  New alt text: $($updateResponse.alt_text)"

    # Step 8: Delete image
    Write-Step "8/8" "Deleting image..."
    Invoke-RestMethod -Uri "$baseUrl$apiPath/$tenantCode/images/$imageId" `
        -Method Delete -Headers $headers | Out-Null

    Write-Success "✓ Image deleted"

    # Cleanup: Delete product
    Write-Info "`nCleaning up..."
    Invoke-RestMethod -Uri "$baseUrl$apiPath/$tenantCode/products/$productId" `
        -Method Delete -Headers $headers | Out-Null
    Write-Success "✓ Test product deleted"

    # Summary
    Write-Host "`n=========================================" -ForegroundColor Green
    Write-Host "✓ ALL TESTS PASSED" -ForegroundColor Green
    Write-Host "=========================================`n" -ForegroundColor Green

    Write-Info "Workflow completed successfully:"
    Write-Info "  ✓ Created test image"
    Write-Info "  ✓ Authenticated user"
    Write-Info "  ✓ Created product"
    Write-Info "  ✓ Uploaded image"
    Write-Info "  ✓ Listed images"
    Write-Info "  ✓ Retrieved image details"
    Write-Info "  ✓ Updated image metadata"
    Write-Info "  ✓ Deleted image"
    Write-Info "  ✓ Cleaned up test data"

    Write-Host "`nNext steps:" -ForegroundColor Cyan
    Write-Host "  1. Check worker logs: make worker-logs"
    Write-Host "  2. Verify image variants were created"
    Write-Host "  3. Check uploads/ directory for files`n"

} catch {
    Write-Host "`n=========================================" -ForegroundColor Red
    Write-Host "✗ TEST FAILED" -ForegroundColor Red
    Write-Host "=========================================`n" -ForegroundColor Red
    Write-Error "Error: $($_.Exception.Message)"
    Write-Host "`nStack Trace:" -ForegroundColor Yellow
    Write-Host $_.ScriptStackTrace
    exit 1
}
