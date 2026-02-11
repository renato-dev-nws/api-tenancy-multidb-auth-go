# CORS Configuration Guide

## Overview
O sistema está configurado com middleware CORS para permitir que aplicações frontend acessem as APIs do backend. A configuração atual permite origens de desenvolvimento e pode ser estendida para produção.

## Configuração Atual

### Origens Permitidas (AllowOrigins)
```go
"http://localhost:3000"   // React/Next.js default
"http://localhost:5173"   // Vite default
"http://localhost:5174"   // Vite alternative 
"http://localhost:8080"   // Vue CLI default
```

### Métodos Permitidos (AllowMethods)
- `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`

### Headers Permitidos (AllowHeaders)
- `Origin`, `Content-Type`, `Accept`, `Authorization`, `X-Requested-With`

### Configuração de Credenciais
- `AllowCredentials: true` - Permite envio de cookies e headers de autenticação
- `MaxAge: 12 hours` - Cache de preflight requests

## Testando o CORS

### 1. Verificar Status dos Servidores
```bash
# Terminal 1 - Admin API
go run ./cmd/admin-api

# Terminal 2 - Tenant API  
go run ./cmd/tenant-api
```

### 2. Teste Basic de CORS
```javascript
// No console do browser (localhost:5174)
fetch('http://localhost:8080/api/health', {
  method: 'GET',
  credentials: 'include'
})
.then(response => response.json())
.then(data => console.log('CORS working:', data))
.catch(error => console.error('CORS error:', error));
```

### 3. Teste de Preflight (OPTIONS)
```javascript
// Teste com headers customizados
fetch('http://localhost:8080/api/health', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer your-token'
  },
  credentials: 'include'
});
```

## Configuração para Produção

### Variáveis de Ambiente
```env
# .env
CORS_ORIGINS=https://app.yourdomain.com,https://admin.yourdomain.com
CORS_CREDENTIALS=true
CORS_MAX_AGE=86400
```

### Implementação Dinâmica
```go
// internal/config/cors.go
func GetCORSConfig() cors.Config {
    origins := os.Getenv("CORS_ORIGINS")
    if origins == "" {
        // Default development origins
        origins = "http://localhost:3000,http://localhost:5173,http://localhost:5174"
    }
    
    return cors.Config{
        AllowOrigins:     strings.Split(origins, ","),
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
}
```

## Troubleshooting

### Erro: "Access to fetch has been blocked by CORS policy"
1. Verificar se a origem do frontend está na lista `AllowOrigins`
2. Confirmar que o servidor backend está rodando
3. Verificar headers nas DevTools → Network

### Preflight Failures
1. Verificar se `OPTIONS` está em `AllowMethods`
2. Confirmar headers permitidos em `AllowHeaders` 
3. Verificar timeout de preflight cache

### Credenciais Rejeitadas
1. Confirmar `AllowCredentials: true`
2. Verificar que origem não é wildcard (`*`)
3. Testar com `credentials: 'include'` no fetch

## Logs para Debug

### Middleware de Log CORS (Opcional)
```go
router.Use(gin.Logger())
router.Use(func(c *gin.Context) {
    origin := c.Request.Header.Get("Origin")
    method := c.Request.Method
    log.Printf("CORS: Origin=%s Method=%s", origin, method)
    c.Next()
})
```

### Verificar Headers de Response
```bash
curl -H "Origin: http://localhost:5174" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: X-Requested-With" \
     -X OPTIONS \
     http://localhost:8080/api/health
```

## Segurança

### Desenvolvimento vs Produção
- **Dev**: Permite `localhost` com portas variadas
- **Prod**: Apenas domínios específicos e HTTPS
- **Staging**: Subdomínios de teste controlados

### Headers de Segurança Adiccionais
```go
// Middleware adicional de segurança
router.Use(func(c *gin.Context) {
    c.Header("X-Frame-Options", "DENY")
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("X-XSS-Protection", "1; mode=block")
    c.Next()
})
```