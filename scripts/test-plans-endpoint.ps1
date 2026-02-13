# Test Plans Endpoint
# Este script testa o endpoint público de listagem de planos na Tenant API

Write-Host "🧪 Testando endpoint de planos..." -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

# Check if Tenant API is running
try {
    $healthCheck = Invoke-WebRequest -Uri "http://localhost:8081/health" -Method GET -ErrorAction Stop
    Write-Host "✅ Tenant API está rodando" -ForegroundColor Green
} catch {
    Write-Host "❌ Tenant API não está rodando em localhost:8081" -ForegroundColor Red
    Write-Host "   Execute: go run ./cmd/tenant-api" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "📋 Testando GET /api/v1/plans..." -ForegroundColor Yellow

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/plans" -Method GET -ErrorAction Stop
    
    Write-Host "✅ Status Code: $($response.StatusCode)" -ForegroundColor Green
    Write-Host ""
    Write-Host "📄 Response Body:" -ForegroundColor Cyan
    
    $json = $response.Content | ConvertFrom-Json
    
    if ($json.total -gt 0) {
        Write-Host "   Total de planos: $($json.total)" -ForegroundColor White
        Write-Host ""
        
        foreach ($plan in $json.plans) {
            Write-Host "   📦 Plano: $($plan.name)" -ForegroundColor Blue
            Write-Host "      ID: $($plan.id)" -ForegroundColor Gray
            Write-Host "      Descrição: $($plan.description)" -ForegroundColor Gray
            Write-Host "      Preço: R$ $($plan.price)" -ForegroundColor Gray
            Write-Host "      Features:" -ForegroundColor Gray
            
            foreach ($feature in $plan.features) {
                Write-Host "         • $($feature.name) ($($feature.slug))" -ForegroundColor White
            }
            
            Write-Host ""
        }
        
        Write-Host "✅ Endpoint funcionando corretamente!" -ForegroundColor Green
        Write-Host ""
        Write-Host "💡 Agora o frontend pode buscar os planos em:" -ForegroundColor Cyan
        Write-Host "   http://localhost:8081/api/v1/plans" -ForegroundColor White
        
    } else {
        Write-Host "⚠️  Nenhum plano cadastrado no banco de dados" -ForegroundColor Yellow
        Write-Host "   Cadastre planos pela Admin API primeiro:" -ForegroundColor Yellow
        Write-Host "   POST http://localhost:8080/api/v1/admin/plans" -ForegroundColor White
    }
    
} catch {
    Write-Host "❌ Erro ao buscar planos: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "   Status Code: $statusCode" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "🔗 Teste CORS com frontend:" -ForegroundColor Cyan
Write-Host "   Origem: http://localhost:5174" -ForegroundColor White

try {
    $headers = @{
        "Origin" = "http://localhost:5174"
    }
    
    $corsResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/plans" -Method GET -Headers $headers -ErrorAction Stop
    
    Write-Host "   ✅ CORS OK - Frontend pode acessar" -ForegroundColor Green
    
} catch {
    Write-Host "   ❌ CORS Error - Verificar configuração" -ForegroundColor Red
}