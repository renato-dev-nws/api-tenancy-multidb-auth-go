# Multi-Tenant SaaS API

API SaaS escalável em Go com isolamento de dados por banco físico (database-per-tenant).

## 🚀 Características

- **Database-per-Tenant**: Cada tenant possui seu próprio banco de dados físico para completo isolamento de dados
- **Control Plane**: Banco Master centralizado para gerenciamento de usuários, tenants, planos e RBAC
- **Feature-Based Plans**: Sistema de planos com features dinâmicas (módulos habilitáveis)
- **RBAC**: Controle de acesso baseado em roles e permissões
- **Connection Pooling**: PgBouncer para gerenciamento eficiente de conexões
- **Cache Layer**: Redis para cache de mapeamentos e mensageria

## 🏗️ Arquitetura

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────┐
│           API Server (Gin)                  │
│  ┌────────────────────────────────────┐    │
│  │  Middleware Stack                  │    │
│  │  • Auth (JWT)                      │    │
│  │  • Tenant Resolution               │    │
│  │  • Feature/Permission Injection    │    │
│  └────────────────────────────────────┘    │
└───┬─────────────┬────────────────┬──────────┘
    │             │                │
┌───▼──────┐  ┌──▼─────────┐  ┌──▼──────┐
│  Master  │  │  Tenant DB │  │  Redis  │
│    DB    │  │ (Dynamic)  │  │  Cache  │
└──────────┘  └────────────┘  └─────────┘
```

### Master DB (Control Plane)
- Users & Profiles
- Tenants & Plans
- Features & Permissions
- RBAC (Roles & Role Permissions)
- Tenant Members

### Tenant DB (Isolated)
- Products
- Services
- Settings
- *Schema aplicado dinamicamente via Worker*

## 📋 Pré-requisitos

- **Go** 1.23+
- **Docker** & Docker Compose
- **Make** (opcional, mas recomendado)

## 🔧 Setup Rápido

### 1. Clone o repositório
```bash
git clone <repository-url>
cd saas-multi-database-api
```

### 2. Configure as variáveis de ambiente
```bash
cp .env.example .env
# Edite .env com suas configurações
```

### 3. Inicie a infraestrutura
```bash
make docker-up
```

Isso irá iniciar:
- PostgreSQL (porta 5432)
- PgBouncer (porta 6432)
- Redis (porta 6379)

### 4. Execute as migrations
```bash
make migrate-up
```

### 5. Inicie a API
```bash
make dev
```

A API estará disponível em `http://localhost:8080`

## 🛠️ Comandos Disponíveis

```bash
make setup          # Configurar ambiente de desenvolvimento
make dev            # Rodar API localmente
make migrate-up     # Executar migrations
make migrate-down   # Reverter migrations
make docker-up      # Iniciar serviços Docker
make docker-down    # Parar serviços Docker
make logs           # Ver logs do Docker
make test           # Executar testes
make build          # Compilar aplicação
make clean          # Limpar artefatos
```

## 📡 Endpoints da API

### Autenticação (Public)

#### Registrar novo usuário
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "full_name": "John Doe"
}
```

#### Login
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:
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
      "id": "tenant-uuid",
      "url_code": "2RT64H3JD77",
      "name": "My Company",
      "role": "admin"
    }
  ]
}
```

### Rotas Protegidas (Requer Autenticação)

#### Obter dados do usuário
```bash
GET /api/v1/auth/me
Authorization: Bearer <token>
```

### Rotas de Tenant (Requer Autenticação + Tenant Access)

#### Obter configuração do tenant (para frontend)
```bash
GET /api/v1/adm/:url_code/config
Authorization: Bearer <token>
```

Response:
```json
{
  "features": ["products", "services"],
  "permissions": ["create_product", "read_product", "update_product"]
}
```

#### Produtos (requer feature 'products')
```bash
# Listar produtos
GET /api/v1/adm/:url_code/products
Authorization: Bearer <token>

# Criar produto (requer permissão 'create_product')
POST /api/v1/adm/:url_code/products
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Product Name",
  "price": 99.99
}
```

#### Serviços (requer feature 'services')
```bash
# Listar serviços
GET /api/v1/adm/:url_code/services
Authorization: Bearer <token>

# Criar serviço (requer permissão 'create_service')
POST /api/v1/adm/:url_code/services
Authorization: Bearer <token>
```

## 🔐 Fluxo de Autenticação e Autorização

### 1. Autenticação (Auth Middleware)
```
Cliente → Header: "Authorization: Bearer <token>"
    ↓
Validação JWT
    ↓
Context: user_id, user_email
```

### 2. Resolução de Tenant (Tenant Middleware)
```
Rota: /api/v1/adm/:url_code/...
    ↓
Extrai url_code do parâmetro
    ↓
Busca db_code no Redis (cache)
    ↓
Se não encontrado → Query Master DB
    ↓
Verifica acesso do usuário (tenant_members)
    ↓
Busca features do plano
    ↓
Busca permissions do usuário
    ↓
Cria/recupera pool do banco tenant
    ↓
Context: tenant_id, tenant_pool, features[], permissions[]
```

### 3. Autorização
```
Feature Check → middleware.RequireFeature("products")
    ↓
Permission Check → middleware.RequirePermission("create_product")
    ↓
Handler executa com acesso ao tenant_pool
```

## 🗄️ Dados Iniciais

O sistema vem com dados de exemplo pré-configurados:

### Features
- `products` - Módulo de produtos
- `services` - Módulo de serviços

### Plans
- **Products Plan** ($19.99) - Acesso apenas ao módulo de produtos
- **Services Plan** ($19.99) - Acesso apenas ao módulo de serviços
- **Premium Plan** ($39.99) - Acesso a todos os módulos (produtos e serviços)

### Permissions
- `create_product`, `read_product`, `update_product`, `delete_product`
- `create_service`, `read_service`, `update_service`, `delete_service`
- `manage_users`, `manage_settings`

### Role Global
- `global_admin` - Acesso a todas as permissões

## 🧪 Testando o Sistema

### 1. Registrar um usuário
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123",
    "full_name": "Admin User"
  }'
```

### 2. Fazer login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

*Nota: Para testar completamente, você precisará criar um tenant e associar o usuário a ele através de queries SQL diretas ou criando endpoints específicos.*

## 📁 Estrutura do Projeto

```
.
├── cmd/
│   └── api/              # Aplicação principal
├── internal/
│   ├── cache/            # Cliente Redis
│   ├── config/           # Configurações
│   ├── database/         # Gerenciador de pools
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # Middlewares (Auth, Tenant)
│   ├── models/           # Modelos de dados
│   ├── repository/       # Camada de acesso a dados
│   └── utils/            # Utilitários (JWT, hash, etc)
├── migrations/
│   ├── master/           # Migrations Master DB
│   └── tenant/           # Migrations Tenant DB
├── config/
│   └── pgbouncer/        # Configuração PgBouncer
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── .env.example
```

## 🔄 Provisionamento de Tenant (TODO)

O sistema está preparado para provisionamento assíncrono:

1. API cria registro no Master DB (`status='provisioning'`)
2. Publica evento no Redis
3. Worker consome evento
4. Worker executa `CREATE DATABASE`
5. Worker aplica migrations
6. Worker atualiza status para `active`

## 📝 Próximos Passos

- [ ] Implementar Worker de provisionamento
- [ ] Adicionar endpoints de gerenciamento de tenants
- [ ] Implementar handlers completos de Products/Services
- [ ] Adicionar testes unitários e de integração
- [ ] Implementar rate limiting
- [ ] Adicionar logging estruturado
- [ ] Implementar métricas e observabilidade

