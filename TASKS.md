# Tasks Roadmap - SaaS Multi-Database API

## 📋 Status Geral
- ✅ **Concluído**: Sistema base de subscription e provisionamento
- 🔄 **Em Progresso**: Implementação de features administrativas e media
- ⏳ **Pendente**: 17 tarefas principais

---

## 🔐 ADMINISTRAÇÃO DO SAAS (Admin API - Porta 8080)

### 1. CRUD de Planos (Plans) ⏳
**Objetivo**: Gerenciar planos do sistema via Admin API

**Endpoints necessários**:
- `GET /api/v1/admin/plans` - Listar todos os planos
- `GET /api/v1/admin/plans/:id` - Obter plano específico
- `POST /api/v1/admin/plans` - Criar novo plano
- `PUT /api/v1/admin/plans/:id` - Atualizar plano
- `DELETE /api/v1/admin/plans/:id` - Deletar plano (soft delete)

**Campos do Plan**:
```go
type PlanRequest struct {
    Name        string  `json:"name" binding:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" binding:"required"`
    IsActive    bool    `json:"is_active"`
}
```

**Features**:
- Associar/desassociar features ao plano
- Listar tenants usando o plano
- Validação: não permitir deletar plano em uso

**Arquivos a criar**:
- `internal/handlers/admin/plan_handler.go`
- `internal/repository/plan_repository.go`
- `internal/services/plan_service.go`

---

### 2. CRUD de Features ⏳
**Objetivo**: Gerenciar features/módulos do sistema

**Endpoints necessários**:
- `GET /api/v1/admin/features` - Listar features
- `GET /api/v1/admin/features/:id` - Obter feature
- `POST /api/v1/admin/features` - Criar feature
- `PUT /api/v1/admin/features/:id` - Atualizar feature
- `DELETE /api/v1/admin/features/:id` - Deletar feature

**Campos da Feature**:
```go
type FeatureRequest struct {
    Title       string `json:"title" binding:"required"`
    Slug        string `json:"slug" binding:"required"` // Ex: products, services
    Code        string `json:"code" binding:"required"` // Ex: prod, serv (para permissões)
    Description string `json:"description"`
    IsActive    bool   `json:"is_active"`
}
```

**Features**:
- Validação de slug único
- Listar planos que usam a feature
- Auto-geração de `code` baseado em slug

**Arquivos a criar**:
- `internal/handlers/admin/feature_handler.go`
- Adicionar método `code` na migration

---

### 3. CRUD de Usuários Admin ⏳
**Objetivo**: Gerenciar usuários do Control Plane (sys_users)

**Endpoints necessários**:
- `GET /api/v1/admin/users` - Listar usuários admin
- `GET /api/v1/admin/users/:id` - Obter usuário
- `POST /api/v1/admin/users` - Criar usuário admin
- `PUT /api/v1/admin/users/:id` - Atualizar usuário
- `DELETE /api/v1/admin/users/:id` - Desativar usuário

**Campos**:
```go
type AdminUserRequest struct {
    Email    string   `json:"email" binding:"required,email"`
    Password string   `json:"password" binding:"required,min=8"` // Apenas criação
    FullName string   `json:"full_name" binding:"required"`
    Status   string   `json:"status"` // active, inactive
    RoleIDs  []int    `json:"role_ids"` // sys_roles
}
```

**Features**:
- RBAC: Associar sys_roles ao usuário
- Não permitir auto-delete
- Logs de auditoria (quem criou/editou)

**Arquivos a criar**:
- `internal/handlers/admin/admin_user_handler.go`
- `internal/repository/admin_user_repository.go`

---

## 🎨 TENANTS - Settings & Layout

### 4. Seed de Settings de Layout ⏳
**Objetivo**: Configurações visuais por tenant

**Campos em `tenant_profiles.custom_settings`**:
```json
{
  "layout": {
    "logo_url": "",
    "primary_color": "#3B82F6",
    "secondary_color": "#10B981",
    "font_family": "Inter",
    "theme": "light"
  }
}
```

**Migration tenant**:
```sql
-- Adicionar defaults na migration tenant
INSERT INTO settings (key, value) VALUES 
  ('logo_url', ''),
  ('primary_color', '#3B82F6'),
  ('secondary_color', '#10B981');
```

**Endpoint**:
- `GET /api/v1/adm/:url_code/settings` - Obter settings
- `PUT /api/v1/adm/:url_code/settings` - Atualizar settings

**Arquivos**:
- `migrations/tenant/002_settings_table.up.sql`
- `internal/handlers/settings_handler.go`

---

## 🔑 TENANTS - Auth Enhancements

### 5-7. Melhorias no Payload de Login ⏳

**Objetivo**: Login response completo com permissões e features

**Novo formato de response**:
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "John Doe"
  },
  "tenants": [
    {
      "id": "uuid",
      "url_code": "27PCKWWWN3F",
      "subdomain": "joao",
      "name": "João Silva",
      "role": "owner",
      "features": ["products", "services"],
      "permissions": ["prod_c", "prod_r", "prod_u", "prod_d", "serv_r"]
    }
  ]
}
```

**Sistema de códigos de permissões**:
| Feature | Code | Permissões |
|---------|------|------------|
| Products | `prod` | `prod_c`, `prod_r`, `prod_u`, `prod_d` |
| Services | `serv` | `serv_c`, `serv_r`, `serv_u`, `serv_d` |
| Users | `user` | `user_c`, `user_r`, `user_u`, `user_d` |
| Settings | `sett` | `sett_r`, `sett_u` |

**Implementação**:
1. Adicionar campo `code` na tabela `features`
2. Atualizar `permissions.slug` para usar formato `{code}_{action}`
3. Criar função `GetUserPermissionCodes(userID, tenantID) []string`
4. Atualizar `TenantLoginResponse` model
5. Middleware `RequirePermission()` deve aceitar códigos

**Arquivos a modificar**:
- `internal/models/models.go` - UserTenant struct
- `internal/repository/tenant_repository.go` - GetUserPermissions
- `internal/handlers/tenant_auth_handler.go` - Login response
- `migrations/master/001_initial_schema.up.sql` - Atualizar permissions

---

## 📦 TENANTS - CRUDs de Domínio

### 8. CRUD de Produtos ⏳

**Migration tenant**:
```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INTEGER DEFAULT 0,
    sku VARCHAR(100) UNIQUE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Endpoints**:
- `GET /api/v1/adm/:url_code/products` - Listar (paginação)
- `GET /api/v1/adm/:url_code/products/:id` - Obter
- `POST /api/v1/adm/:url_code/products` - Criar (requer `prod_c`)
- `PUT /api/v1/adm/:url_code/products/:id` - Atualizar (requer `prod_u`)
- `DELETE /api/v1/adm/:url_code/products/:id` - Deletar (requer `prod_d`)

**Features**:
- Paginação (limit, offset)
- Filtros (search, is_active)
- Ordenação (price, name, created_at)

**Arquivos**:
- `migrations/tenant/003_products_table.up.sql`
- `internal/models/product.go`
- `internal/handlers/product_handler.go`
- `internal/repository/product_repository.go`

---

### 9. CRUD de Serviços ⏳

**Migration tenant**:
```sql
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    duration_minutes INTEGER, -- Duração do serviço
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Endpoints**: Similar ao Products, substituindo `/products` por `/services`

**Features**:
- Campo específico: `duration_minutes`
- Permissões: `serv_c`, `serv_r`, `serv_u`, `serv_d`

**Arquivos**:
- `migrations/tenant/004_services_table.up.sql`
- `internal/models/service.go`
- `internal/handlers/service_handler.go`
- `internal/repository/service_repository.go`

---

## 🖼️ SISTEMA DE MEDIA (Tenant DB)

### 10-11. Tabela de Imagens (Polymorphic) ⏳

**Migration tenant**:
```sql
CREATE TYPE media_type AS ENUM ('image', 'video', 'document');
CREATE TYPE image_variant AS ENUM ('original', 'medium', 'small', 'thumb');

CREATE TABLE images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Polymorphic relationship
    imageable_type VARCHAR(50) NOT NULL, -- 'product', 'service', 'user', etc
    imageable_id UUID NOT NULL,
    
    -- File metadata
    filename VARCHAR(255) NOT NULL, -- UUID-based filename
    original_filename VARCHAR(255), -- Nome original do upload
    title VARCHAR(255), -- Editável pelo usuário
    alt_text VARCHAR(255),
    
    -- Media info
    media_type media_type NOT NULL DEFAULT 'image',
    mime_type VARCHAR(100) NOT NULL, -- image/webp
    extension VARCHAR(10) NOT NULL, -- webp
    
    -- Image variant
    variant image_variant NOT NULL DEFAULT 'original',
    parent_id UUID REFERENCES images(id) ON DELETE CASCADE, -- Referência à original
    
    -- Dimensions
    width INTEGER,
    height INTEGER,
    file_size BIGINT, -- bytes
    
    -- Storage
    storage_driver VARCHAR(20) NOT NULL DEFAULT 'local', -- local, s3, r2
    storage_path TEXT NOT NULL, -- /uploads/{tenant}/images/products/{id}/{uuid}_original.webp
    public_url TEXT, -- URL de acesso público
    
    -- Processing
    processing_status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    processed_at TIMESTAMP,
    
    -- Order
    display_order INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Indexes
    INDEX idx_images_imageable (imageable_type, imageable_id),
    INDEX idx_images_variant (variant),
    INDEX idx_images_parent (parent_id)
);
```

**Model Go**:
```go
type Image struct {
    ID              uuid.UUID     `json:"id"`
    ImageableType   string        `json:"imageable_type"`
    ImageableID     uuid.UUID     `json:"imageable_id"`
    Filename        string        `json:"filename"`
    OriginalFilename string       `json:"original_filename,omitempty"`
    Title           string        `json:"title,omitempty"`
    AltText         string        `json:"alt_text,omitempty"`
    MediaType       string        `json:"media_type"`
    MimeType        string        `json:"mime_type"`
    Extension       string        `json:"extension"`
    Variant         string        `json:"variant"`
    ParentID        *uuid.UUID    `json:"parent_id,omitempty"`
    Width           int           `json:"width"`
    Height          int           `json:"height"`
    FileSize        int64         `json:"file_size"`
    StorageDriver   string        `json:"storage_driver"`
    StoragePath     string        `json:"storage_path"`
    PublicURL       string        `json:"public_url"`
    ProcessingStatus string       `json:"processing_status"`
    ProcessedAt     *time.Time    `json:"processed_at,omitempty"`
    DisplayOrder    int           `json:"display_order"`
    CreatedAt       time.Time     `json:"created_at"`
    UpdatedAt       time.Time     `json:"updated_at"`
}
```

---

### 12-13. Upload de Imagens ⏳

**Endpoint**:
```bash
POST /api/v1/adm/:url_code/images
Content-Type: multipart/form-data

Form fields:
- files[]: File[] (até 10 arquivos)
- imageable_type: string (product, service)
- imageable_id: uuid
- titles[]: string[] (opcional)
```

**Response**:
```json
{
  "uploaded": 3,
  "images": [
    {
      "id": "uuid",
      "filename": "550e8400-e29b-41d4-a716-446655440000_original.webp",
      "storage_path": "/uploads/tenant-uuid/images/products/product-uuid/...",
      "processing_status": "pending"
    }
  ]
}
```

**Validações**:
- Máximo 10 arquivos por request
- Tipos permitidos: jpg, jpeg, png, webp
- Tamanho máximo por arquivo: 10MB
- Validar que `imageable_id` existe

**Estrutura de diretórios**:
```
uploads/
└── {tenant_uuid}/
    └── images/
        ├── products/
        │   └── {product_uuid}/
        │       ├── {image_uuid}_original.webp
        │       ├── {image_uuid}_medium.webp
        │       ├── {image_uuid}_small.webp
        │       └── {image_uuid}_thumb.webp
        └── services/
            └── {service_uuid}/
                └── ...
```

**Arquivos**:
- `internal/handlers/image_handler.go`
- `internal/services/upload_service.go`
- `internal/storage/storage.go` (interface)
- `internal/storage/local.go`
- Criar pasta `uploads/` no root do projeto
- Servir arquivos estáticos via Gin: `router.Static("/uploads", "./uploads")`

---

### 14-16. Worker de Processamento de Imagens ⏳

**Fluxo**:
1. Upload → Salva original no disco
2. Cria registro na tabela `images` (status: pending)
3. Publica evento no Redis: `image:process:{tenant_uuid}:{image_id}`
4. Worker consome evento
5. Worker redimensiona usando ImageMagick (ou Go lib: `imaging`)
6. Cria 3 variantes (medium, small, thumb)
7. Salva registros filhos na tabela `images`
8. Atualiza status original para `completed`

**Tamanhos de redimensionamento**:
```go
const (
    SizeOriginal = "1400x800"   // max width x max height
    SizeMedium   = "800x600"
    SizeSmall    = "350x200"
    SizeThumb    = "100x100"
)
```

**Library recomendada**: `github.com/disintegration/imaging`
```go
import "github.com/disintegration/imaging"

// Redimensiona mantendo aspect ratio
img := imaging.Fit(src, 800, 600, imaging.Lanczos)

// Salva como WebP
imaging.Save(img, outputPath, imaging.WebP)
```

**Redis Queue**:
```go
type ImageProcessEvent struct {
    TenantID  uuid.UUID `json:"tenant_id"`
    ImageID   uuid.UUID `json:"image_id"`
    DBCode    string    `json:"db_code"`
}
```

**Arquivos**:
- `cmd/image-worker/main.go` (novo worker)
- `internal/services/image_processor.go`
- Adicionar ao `docker-compose.yml`

---

### 17. Configuração Multi-Storage ⏳

**Variáveis `.env`**:
```env
# Storage Configuration
STORAGE_DRIVER=local # local, s3, r2

# Local Storage
UPLOADS_PATH=./uploads

# AWS S3
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=us-east-1
AWS_BUCKET=my-saas-uploads

# Cloudflare R2
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_ACCOUNT_ID=
R2_BUCKET=my-saas-uploads
R2_PUBLIC_URL=https://pub-xxxxx.r2.dev
```

**Interface de Storage**:
```go
type StorageDriver interface {
    Upload(ctx context.Context, file io.Reader, path string) (string, error)
    Delete(ctx context.Context, path string) error
    GetPublicURL(path string) string
}

type LocalStorage struct{}
type S3Storage struct{}
type R2Storage struct{}
```

**Factory**:
```go
func NewStorageDriver(driver string) StorageDriver {
    switch driver {
    case "s3":
        return &S3Storage{}
    case "r2":
        return &R2Storage{}
    default:
        return &LocalStorage{}
    }
}
```

**Arquivos**:
- `internal/storage/interface.go`
- `internal/storage/local.go`
- `internal/storage/s3.go`
- `internal/storage/r2.go`
- `internal/storage/factory.go`

---

## 📊 Resumo de Implementação

### Ordem Sugerida:

**Fase 1: Admin API (Base)** ⏳
1. CRUD Planos
2. CRUD Features (adicionar campo `code`)
3. CRUD Usuários Admin

**Fase 2: Auth & Permissions** ⏳
4. Atualizar permissions com códigos
5. Melhorar payload de login
6. Seed de settings de layout

**Fase 3: Domain CRUDs** ⏳
7. CRUD Produtos
8. CRUD Serviços

**Fase 4: Media System** ⏳
9. Tabela images (migration)
10. Model e repository
11. Upload endpoint (local)
12. Worker de processamento
13. Redimensionamento e WebP
14. Multi-storage (S3/R2)

---

## 🛠️ Tecnologias Necessárias

**Dependências Go**:
```bash
go get github.com/disintegration/imaging  # Image processing
go get github.com/aws/aws-sdk-go-v2       # S3
# R2 usa S3-compatible API
```

**Docker**:
- Adicionar `image-worker` ao compose
- Volume compartilhado para `./uploads`

**Libraries frontend sugeridas**:
- Upload: `react-dropzone`
- Preview: `react-image-gallery`

---

## ✅ Checklist de Conclusão

Cada tarefa deve incluir:
- [ ] Migration SQL
- [ ] Model Go
- [ ] Repository (CRUD methods)
- [ ] Service (business logic)
- [ ] Handler (HTTP endpoints)
- [ ] Middleware (permissions)
- [ ] Tests unitários
- [ ] Documentação da API (README/Postman)
- [ ] Commit com mensagem descritiva

---

**Última atualização**: 2026-02-07
**Total de tarefas**: 17
**Estimativa total**: 4-6 semanas
