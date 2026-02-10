# Image Processing Worker

## 📋 Visão Geral

Worker assíncrono responsável por processar imagens enviadas pelos usuários, gerando variantes otimizadas em diferentes tamanhos. O processamento acontece em background para não bloquear o upload e melhorar a experiência do usuário.

## 🏗️ Arquitetura

### Componentes Principais

```
┌─────────────────┐      ┌─────────────┐      ┌──────────────────┐
│  Tenant API     │─────▶│   Redis     │─────▶│  Image Worker    │
│  (Upload)       │      │  Pub/Sub    │      │  (Processing)    │
└─────────────────┘      └─────────────┘      └──────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │  Storage Layer  │
                                               │  (Local/S3/R2)  │
                                               └─────────────────┘
```

### Fluxo de Dados

1. **Upload**: Tenant API recebe imagem via multipart/form-data
2. **Persistência**: Imagem original salva no storage (status: `pending`)
3. **Evento Redis**: Publica evento no canal `image:process`
4. **Consumo**: Worker recebe evento da fila
5. **Processamento**: Gera variantes redimensionadas
6. **Finalização**: Atualiza status para `completed` ou `failed`

## 🔄 Fluxo de Processamento Detalhado

### 1. Recebimento do Evento

```json
{
  "tenant_db_code": "550e8400-e29b-41d4-a716-446655440000",
  "image_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Canal Redis**: `image:process`

### 2. Validação Inicial

```go
// Verifica se imagem está elegível para processamento
if img.Variant != VariantOriginal || img.ProcessingStatus != StatusPending {
    return error
}
```

**Critérios**:
- ✅ Deve ser variante `original`
- ✅ Status deve ser `pending`
- ❌ Rejeita imagens já processadas ou em processamento

### 3. Atualização de Status

```
pending → processing
```

Previne processamento duplicado em caso de múltiplos workers.

### 4. Download da Imagem Original

```go
reader, err := storageDriver.GetReader(ctx, img.StoragePath)
srcImage, format, err := image.Decode(reader)
```

**Formatos Suportados**:
- JPEG/JPG
- PNG
- GIF
- WebP (futuro)

### 5. Extração de Dimensões

```go
bounds := srcImage.Bounds()
originalWidth := bounds.Dx()   // Largura
originalHeight := bounds.Dy()  // Altura
```

Atualiza registro no banco com dimensões reais.

### 6. Geração de Variantes

Para cada variante configurada:

| Variante   | Max Width | Max Height | Uso                          |
|-----------|-----------|------------|------------------------------|
| `original`| 1400px    | 1400px     | Imagem completa              |
| `medium`  | 800px     | 800px      | Galeria desktop              |
| `small`   | 350px     | 350px      | Galeria mobile, thumbnails   |
| `thumb`   | 100px     | 100px      | Listagens, avatares          |

**Algoritmo de Resize**: Lanczos (melhor qualidade)
**Aspect Ratio**: Mantido (fit dentro das dimensões max)

### 7. Encoding de Variantes

```go
switch format {
case "jpeg", "jpg":
    jpeg.Encode(file, resizedImage, &jpeg.Options{Quality: 90})
case "png":
    png.Encode(file, resizedImage)
default:
    jpeg.Encode(file, resizedImage, &jpeg.Options{Quality: 90})
}
```

**Qualidade JPEG**: 90% (balanço entre qualidade e tamanho)

### 8. Upload das Variantes

```go
// Estrutura de diretórios
{tenant_uuid}/images/{imageable_type}/{imageable_id}/{filename}_{variant}.{ext}

// Exemplo
550e8400/images/product/abc123/photo_medium.jpg
550e8400/images/product/abc123/photo_small.jpg
550e8400/images/product/abc123/photo_thumb.jpg
```

### 9. Persistência no Banco

Para cada variante:
- Cria registro em `images` table
- Define `parent_id` apontando para a imagem original
- Armazena dimensões reais (width, height)
- Registra tamanho do arquivo (file_size)
- Define status como `completed`

### 10. Finalização

```
processing → completed  (sucesso)
processing → failed     (erro)
```

**Em caso de erro**:
- Cleanup automático: remove arquivos já enviados
- Log detalhado do erro
- Status atualizado para `failed`
- Imagem pode ser reprocessada manualmente

## ⚙️ Configuração

### Variáveis de Ambiente

```env
# Database (Master DB via PgBouncer)
MASTER_DB_HOST=pgbouncer
MASTER_DB_PORT=5432
MASTER_DB_USER=saas_api
MASTER_DB_PASSWORD=saas_api_password
MASTER_DB_NAME=master_db
MASTER_DB_SSLMODE=disable

# Direct Postgres (para operações admin)
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=master_db

# Redis (Pub/Sub)
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=""
REDIS_DB=0

# Storage
STORAGE_DRIVER=local          # local | s3 | r2
UPLOADS_PATH=./uploads

# Ambiente
APP_ENV=development           # development | production
```

### Dependências

```bash
# Processamento de imagens
go get github.com/disintegration/imaging@v1.6.2

# Database
go get github.com/jackc/pgx/v5

# Cache/Queue
go get github.com/redis/go-redis/v9
```

## 🚀 Execução

### Local (Desenvolvimento)

```bash
# Via Makefile
make run-image-worker

# Direto
go run cmd/image-worker/main.go
```

### Docker (Produção)

```bash
# Build da imagem
docker build -f Dockerfile.image-worker -t saas-image-worker .

# Execução standalone
docker run --name image-worker \
  --env-file .env \
  -v $(pwd)/uploads:/app/uploads \
  saas-image-worker

# Via Docker Compose
docker compose up -d image-worker
```

### Kubernetes (Escala)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: image-worker
spec:
  replicas: 3  # Múltiplos workers
  template:
    spec:
      containers:
      - name: worker
        image: saas-image-worker:latest
        envFrom:
        - configMapRef:
            name: app-config
        volumeMounts:
        - name: uploads
          mountPath: /app/uploads
```

## 🔍 Monitoramento

### Logs

```bash
# Tempo real
make worker-logs

# Docker
docker compose logs -f image-worker

# Filtrar erros
docker compose logs image-worker | grep "Erro"
```

**Formato dos Logs**:
```
2026/02/10 14:30:45 Processando imagem: tenant=550e8400, image_id=123e4567
2026/02/10 14:30:46 Imagem 123e4567 processada com sucesso
```

### Health Check

```bash
# Verificar se worker está consumindo eventos
redis-cli PUBSUB NUMSUB image:process

# Verificar fila de pending images
psql -c "SELECT COUNT(*) FROM images WHERE processing_status = 'pending'"
```

### Métricas (Sugeridas)

- **Throughput**: Imagens processadas/minuto
- **Latência**: Tempo médio de processamento
- **Taxa de erro**: % de imagens com status `failed`
- **Backlog**: Quantidade de imagens `pending`

## 🛠️ Troubleshooting

### Worker não está consumindo eventos

```bash
# 1. Verificar conexão Redis
redis-cli PING

# 2. Verificar subscriber ativo
redis-cli PUBSUB CHANNELS

# 3. Verificar logs do worker
docker compose logs image-worker --tail 50
```

### Imagens ficam em status "processing"

**Causa**: Worker morreu durante processamento

**Solução**:
```sql
-- Resetar status para reprocessamento
UPDATE images 
SET processing_status = 'pending', 
    updated_at = CURRENT_TIMESTAMP 
WHERE processing_status = 'processing' 
  AND updated_at < NOW() - INTERVAL '1 hour';
```

### Erro "failed to get image reader"

**Causas possíveis**:
1. Arquivo não existe no storage
2. Permissões incorretas no diretório `uploads/`
3. Storage path incorreto no banco

**Debug**:
```bash
# Verificar arquivo existe
ls -lh uploads/{tenant_uuid}/images/...

# Verificar permissões
chmod -R 755 uploads/
```

### Erro de memória (OOM)

**Causa**: Imagens muito grandes sobrecarregam memória

**Soluções**:
1. Aumentar limite de memória do container
2. Implementar streaming (processar em chunks)
3. Limitar tamanho máximo de upload (10MB padrão)

```dockerfile
# Limitar memória no Docker
docker run -m 512m saas-image-worker
```

## 📊 Estatísticas de Processamento

### Benchmarks (Intel i7, 16GB RAM)

| Tamanho Original | Formato | Tempo Total | Variantes |
|-----------------|---------|-------------|-----------|
| 2MB (3000x2000) | JPEG    | ~1.2s       | 3         |
| 5MB (5000x3000) | PNG     | ~3.5s       | 3         |
| 1MB (2000x1500) | JPEG    | ~0.8s       | 3         |

**Estimativa**: ~1.5s por imagem em média

### Capacidade

- **Single Worker**: ~40 imagens/minuto (média)
- **3 Workers**: ~120 imagens/minuto
- **Bottleneck**: I/O de storage (não CPU)

## 🔐 Segurança

### Isolamento de Tenants

- ✅ Cada tenant tem pool de conexão isolado
- ✅ Imagens organizadas por `tenant_uuid`
- ✅ Worker acessa apenas DB do tenant do evento
- ✅ Não há cross-tenant contamination possível

### Validação de Imagens

```go
// Realizado no upload (Tenant API)
- Tamanho máximo: 10MB
- Extensões permitidas: .jpg, .jpeg, .png, .webp, .gif
- MIME type validation
```

### Cleanup de Falhas

```go
// Rollback automático em caso de erro
if err := imageRepo.Create(ctx, variantReq); err != nil {
    storageDriver.Delete(ctx, finalPath)  // Remove arquivo órfão
    return err
}
```

## 🚧 Roadmap

### Futuras Melhorias

- [ ] **WebP Conversion**: Gerar variantes WebP para economia de banda
- [ ] **Batch Processing**: Processar múltiplas imagens em paralelo
- [ ] **Dead Letter Queue**: Retry automático com backoff exponencial
- [ ] **Progress Tracking**: WebSocket para status em tempo real
- [ ] **Watermarking**: Aplicar marca d'água configurável
- [ ] **Smart Cropping**: Detecção de faces/objetos para crops inteligentes
- [ ] **CDN Integration**: Envio direto para CloudFlare/AWS CloudFront
- [ ] **Metrics/Prometheus**: Exportar métricas para monitoramento

## 📚 Referências

- **Imaging Library**: https://github.com/disintegration/imaging
- **Redis Pub/Sub**: https://redis.io/docs/manual/pubsub/
- **Go Image Package**: https://pkg.go.dev/image
- **JPEG Encoding**: https://pkg.go.dev/image/jpeg

## 🤝 Contribuindo

Ao adicionar novas features de processamento:

1. Mantenha variantes configuráveis em `internal/models/tenant/image.go`
2. Adicione testes unitários para novos algoritmos
3. Documente impacto em performance
4. Verifique compatibilidade com todos formatos de imagem
5. Implemente cleanup adequado em caso de erros

---

**Última atualização**: Fevereiro 2026  
**Versão**: 1.0.0  
**Maintainer**: SaaS Team
