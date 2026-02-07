# Multi-Tenant SaaS API

API SaaS escalável em Go com isolamento de dados por banco físico (database-per-tenant).

## 🚀 Características

- **Database-per-Tenant**: Cada tenant possui seu próprio banco de dados físico para completo isolamento de dados
- **Control Plane**: Banco Master centralizado para gerenciamento de usuários, tenants, planos e RBAC
- **Login Inteligente**: Sistema de autenticação que retorna configuração completa do tenant em uma única chamada
- **Interface Dinâmica**: Frontend recebe layout, features e permissões automaticamente
- **Tenant Switching**: Troca de tenant sem novo login, apenas atualizando configurações
- **Feature-Based Plans**: Sistema de planos com features dinâmicas (módulos habilitáveis)
- **RBAC**: Controle de acesso baseado em roles e permissões
- **Dual Routing**: Subdomain para site público + URL code para admin panel
- **Auto-Provisioning**: Worker assíncrono para criação automática de bancos tenant
- **Subscription System**: Endpoint público para auto-cadastro de novos clientes
- **Billing Cycles**: Suporte a faturamento mensal, trimestral, semestral e anual
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
- Tenants & Plans (com billing_cycle)
- Features & Permissions
- RBAC (Roles & Role Permissions)
- Tenant Members

### Tenant DB (Isolated)
- Products
- Services
- Settings
- *Schema aplicado automaticamente via Worker*

### Tenant Routing
- **subdomain**: Escolhido pelo usuário para site público (ex: `joao.meusaas.app`)
- **url_code**: Gerado automaticamente (11 chars A-Z0-9) para admin panel (ex: `meusaas.app/adm/27PCKWWWN3F/dashboard`)

**Exemplo de tenant:**
```
Tenant: João Silva
├─ subdomain: "joao"
│  └─ Site público: https://joao.meusaas.app
└─ url_code: "27PCKWWWN3F" (auto-gerado)
   └─ Admin panel: https://meusaas.app/adm/27PCKWWWN3F/dashboard
```

## � Fluxo de Subscription (Auto-Cadastro)

```
Cliente (Frontend)
    │
    ├─ POST /api/v1/subscription
    │  {
    │    plan_id, billing_cycle, name, email, password,
    │    subdomain: "joao" (escolhido pelo usuário)
    │  }
    │
    ▼
Tenant API
    │
    ├─ Valida dados
    ├─ Hash de senha (bcrypt)
    ├─ Gera url_code: "27PCKWWWN3F" (auto)
    │
    ├─ TRANSACTION BEGIN
    │   ├─ Cria User
    │   ├─ Cria UserProfile
    │   ├─ Cria Tenant (status: provisioning)
    │   ├─ Cria TenantProfile
    │   └─ Adiciona User como Owner
    ├─ TRANSACTION COMMIT
    │
    ├─ Publica evento Redis: "tenant:provision:{db_code}"
    │
    └─ Retorna: { token, user, tenant }
    
    ▼
Worker (Background)
    │
    ├─ Consome evento da fila
    ├─ CREATE DATABASE db_tenant_{db_code}
    ├─ Aplica migrations (schema tenant)
    ├─ UPDATE tenants SET status='active'
    │
    └─ Tenant pronto! (2-5 segundos)
```

## �📋 Pré-requisitos

- **Go** 1.23+
- **Docker** & Docker Compose
- **Make** (opcional, mas recomendado)

## 🔧 Setup Rápido

### 1. Clone o repositório
```bash
git clone <repository-url>
cd saas-multi-database-api
```

### 2. Setup completo (um comando)
```bash
make setup
```

Isso irá:
- ✅ Construir as imagens Docker (Admin API, Tenant API, Worker)
- ✅ Iniciar serviços (PostgreSQL, PgBouncer, Redis)
- ✅ Aplicar migrations no Master DB
- ✅ Criar usuário admin (`admin@teste.com` / `admin123`)

**Serviços iniciados:**
- **Admin API**: http://localhost:8080
- **Tenant API**: http://localhost:8081
- **PostgreSQL**: porta 5432
- **PgBouncer**: porta 6432
- **Redis**: porta 6379

### 3. Testar o sistema
```bash
make test-subscription
```

## ⚙️ Configuração

### Variáveis de Ambiente (Docker Compose)

O sistema está configurado para funcionar out-of-the-box. Principais variáveis:

**PostgreSQL**
- `POSTGRES_USER=postgres`
- `POSTGRES_PASSWORD=postgres`
- `POSTGRES_DB=master_db`

**APIs**
- `ADMIN_API_PORT=8080`
- `TENANT_API_PORT=8081`
- `JWT_SECRET=your-secret-key` (⚠️ mudar em produção)

**Redis**
- `REDIS_HOST=redis:6379`
- `REDIS_QUEUE=tenant:provision`

**PgBouncer**
- Pool mode: `transaction`
- Max connections: `100`

Para customizar, edite `docker-compose.yml` ou crie arquivo `.env`.

## 🛠️ Comandos Disponíveis

```bash
# Setup
make setup               # Setup completo (build + migrate + seed)
make reset               # Reset total (down -v + setup)
make start               # Iniciar serviços
make stop                # Parar serviços
make restart             # Reiniciar serviços

# Development
make logs                # Ver todos os logs
make logs-admin          # Logs da Admin API
make logs-tenant         # Logs da Tenant API
make logs-worker         # Logs do Worker
make migrate             # Aplicar migrations Master DB
make seed                # Criar admin user

# Testing
make test-subscription   # Testar cadastro público
make test-login          # Testar login (retorna interface)
make test-switch-tenant  # Testar troca de tenant ativo
make test-plans-list     # Testar listagem de planos (Admin API)

# Utilities
make clean               # Limpar volumes e rebuild
```

## 📡 Endpoints da API

### 🌐 Subscription (Public - Porta 8081)

#### Cadastro de novo assinante (público)
```bash
POST /api/v1/subscription
Content-Type: application/json

{
  "plan_id": "33333333-3333-3333-3333-333333333333",
  "billing_cycle": "monthly",
  "name": "João Silva",
  "is_company": false,
  "company_name": "Minha Empresa Ltda",  // Opcional se is_company=false
  "subdomain": "joao",
  "email": "joao@teste.com",
  "password": "senha12345",
  "custom_domain": "app.minhaempresa.com"  // Opcional
}
```

**Billing Cycles**: `monthly`, `quarterly`, `semiannual`, `annual`

**Plans Disponíveis**:
- `11111111-1111-1111-1111-111111111111` - Products Plan ($19.99)
- `22222222-2222-2222-2222-222222222222` - Services Plan ($19.99)
- `33333333-3333-3333-3333-333333333333` - Premium Plan ($39.99)

Response:
```json
{
  "token": "eyJhbGc...",
  "current_tenant": {
    "id": "tenant-uuid",
    "url_code": "27PCKWWWN3F",
    "subdomain": "joao",
    "name": "Empresa João Silva"
  },
  "interface": {
    "company_name": "Empresa João Silva",
    "logo_url": "https://cdn.example.com/logo.png",
    "custom_settings": {
      "primary_color": "#3B82F6",
      "theme": "light"
    }
  },
  "features": ["products", "services"],
  "permissions": ["create_product", "read_product", "update_product", "delete_product"]
}
```

### 🔐 Autenticação (Porta 8081)

#### Login Tenant
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "joao@teste.com",
  "password": "senha12345"
}
```

Response:
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "uuid",
    "email": "joao@teste.com",
    "full_name": "João Silva"
  },
  "tenants": [
    {
      "id": "tenant-uuid",
      "url_code": "27PCKWWWN3F",
      "subdomain": "joao",
      "name": "Empresa João Silva",
      "role": "owner"
    }
  ],
  "current_tenant": {
    "id": "tenant-uuid",
    "url_code": "27PCKWWWN3F",
    "subdomain": "joao",
    "name": "Empresa João Silva"
  },
  "interface": {
    "company_name": "Empresa João Silva",
    "custom_settings": {
      "industry": "",
      "name": "Empresa João Silva"
    }
  },
  "features": ["products", "services"],
  "permissions": ["create_product", "read_product", "update_product", "delete_product"]
}
```

### Rotas Protegidas (Requer Autenticação)

#### Obter dados do usuário
```bash
GET /api/v1/auth/me
Authorization: Bearer <token>
```

#### Trocar tenant ativo
```bash
POST /api/v1/auth/switch-tenant
Authorization: Bearer <token>
Content-Type: application/json

{
  "url_code": "27PCKWWWN3F"
}
```

Response:
```json
{
  "message": "tenant switched successfully",
  "current_tenant": {
    "id": "tenant-uuid",
    "url_code": "27PCKWWWN3F",
    "subdomain": "joao",
    "name": "Empresa João Silva"
  },
  "interface": {
    "company_name": "Empresa João Silva",
    "custom_settings": {}
  },
  "features": ["products", "services"],
  "permissions": ["create_product", "read_product"]
}
```

### Rotas de Tenant (Requer Autenticação + Tenant Access)

#### Obter configuração do tenant (para frontend)
```bash
GET /api/v1/:url_code/config
Authorization: Bearer <token>
```

Response:
```json
{
  "features": ["products", "services"],
  "permissions": ["create_product", "read_product", "update_product"],
  "config": {
    "logo_url": "https://cdn.example.com/logo.png",
    "company_name": "Empresa João Silva",
    "custom_settings": {}
  }
}
```

#### Produtos (requer feature 'products')
```bash
# Listar produtos
GET /api/v1/:url_code/products
Authorization: Bearer <token>

# Criar produto (requer permissão 'create_product')
POST /api/v1/:url_code/products
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
GET /api/v1/:url_code/services
Authorization: Bearer <token>

# Criar serviço (requer permissão 'create_service')
POST /api/v1/:url_code/services
Authorization: Bearer <token>
```

### 📊 Resumo de Endpoints

| Método | Endpoint | Auth | Descrição |
|--------|----------|------|-----------|
| `POST` | `/api/v1/subscription` | ❌ Público | Cadastro de novo assinante |
| `POST` | `/api/v1/auth/login` | ❌ Público | Login tenant (retorna interface) |
| `POST` | `/api/v1/auth/switch-tenant` | ✅ JWT | Trocar tenant ativo |
| `GET` | `/api/v1/auth/me` | ✅ JWT | Dados do usuário logado |
| `GET` | `/api/v1/:url_code/config` | ✅ JWT + Tenant | Config do frontend |
| `GET` | `/api/v1/:url_code/products` | ✅ JWT + Feature | Listar produtos |
| `POST` | `/api/v1/:url_code/products` | ✅ JWT + Permission | Criar produto |
| `GET` | `/api/v1/:url_code/services` | ✅ JWT + Feature | Listar serviços |
| `POST` | `/api/v1/:url_code/services` | ✅ JWT + Permission | Criar serviço |
| `POST` | `/api/v1/admin/login` | ❌ Público | Login admin (porta 8080) |
| `POST` | `/api/v1/admin/tenants` | ✅ Admin JWT | Criar tenant (admin) |

**Legenda:**
- ✅ JWT: Requer header `Authorization: Bearer <token>`
- ✅ JWT + Tenant: Requer acesso ao tenant via `tenant_members`
- ✅ JWT + Feature: Requer feature habilitada no plano
- ✅ JWT + Permission: Requer permissão específica do usuário

## 🔐 Fluxo de Autenticação e Autorização

### 1. Login Direto com Interface
```
Cliente → POST /api/v1/auth/login {email, password}
    ↓
Validação credenciais
    ↓
Busca last_tenant_logged do usuário
    ↓
Se tem tenant ativo:
  │
  ├─ Busca configuração do tenant
  ├─ Busca features do plano
  ├─ Busca permissions do usuário
  └─ Busca interface/layout config
    ↓
Retorna: {
  token, user, tenants[],
  current_tenant, interface,
  features[], permissions[]
}
```

### 2. Troca de Tenant (Switch)
```
Cliente → POST /api/v1/auth/switch-tenant {url_code}
    ↓
Auth Middleware → Valida JWT
    ↓
Verifica acesso do usuário ao tenant
    ↓
Atualiza last_tenant_logged
    ↓
Busca nova configuração:
  ├─ Features do novo tenant
  ├─ Permissions do usuário
  └─ Interface/layout config
    ↓
Retorna nova configuração completa
```

### 3. Rotas de Tenant (Resolução Automática)
```
Rota: /api/v1/:url_code/...
    ↓
Auth Middleware → Valida JWT
    ↓
Tenant Middleware:
  ├─ Extrai url_code do parâmetro
  ├─ Busca db_code no Redis (cache)
  ├─ Se não encontrado → Query Master DB
  ├─ Verifica acesso do usuário (tenant_members)
  ├─ Busca features do plano
  ├─ Busca permissions do usuário
  └─ Cria/recupera pool do banco tenant
    ↓
Context: tenant_id, tenant_pool, features[], permissions[]
```

### 4. Autorização
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

### Plans (UUIDs Fixos)
- **11111111-1111-1111-1111-111111111111** - Products Plan ($19.99) - Apenas produtos
- **22222222-2222-2222-2222-222222222222** - Services Plan ($19.99) - Apenas serviços
- **33333333-3333-3333-3333-333333333333** - Premium Plan ($39.99) - Todos os módulos

### Features (UUIDs Fixos)
- **aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa** - products
- **bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb** - services

### Permissions
- `create_product`, `read_product`, `update_product`, `delete_product`
- `create_service`, `read_service`, `update_service`, `delete_service`
- `manage_users`, `manage_settings`

### Tenant Roles
- **owner** - Criador do tenant, acesso total
- **admin** - Administrador, acesso a todas as features habilitadas
- **manager** - Gerente, acesso limitado
- **user** - Usuário comum, acesso somente leitura

### System Roles (Admin API)
- **super_admin** - Acesso total ao sistema
- **admin** - Administrador do Control Plane
- **support** - Suporte técnico
- **viewer** - Visualização apenas

## 🧪 Testando o Sistema

### 1. Cadastro público de assinante
```bash
make test-subscription
```

Ou manualmente:
```bash
curl -X POST http://localhost:8081/api/v1/subscription \
  -H "Content-Type: application/json" \
  -d '{
    "plan_id": "33333333-3333-3333-3333-333333333333",
    "billing_cycle": "monthly",
    "name": "João Silva",
    "is_company": false,
    "subdomain": "joao",
    "email": "joao@teste.com",
    "password": "senha12345"
  }'
```

### 2. Verificar tenant no banco
```bash
docker exec saas-postgres psql -U postgres -d master_db \
  -c "SELECT url_code, subdomain, billing_cycle, status FROM tenants;"
```

### 3. Fazer login e receber configuração completa
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "joao@teste.com",
    "password": "senha12345"
  }'
```

**Response inclui**:
- `current_tenant`: Tenant ativo (baseado em last_tenant_logged)
- `interface`: Configuração de layout (logo, company_name, custom_settings)
- `features`: Features disponíveis no plano ["products", "services"]
- `permissions`: Permissões do usuário no tenant

### 4. Trocar de tenant (se usuário tiver múltiplos)
```bash
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"joao@teste.com","password":"senha12345"}' | \
  grep -o '"token":"[^"]*' | cut -d'"' -f4)

curl -X POST http://localhost:8081/api/v1/auth/switch-tenant \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url_code":"OUTRO_TENANT"}'
```

### 5. Acessar rotas do tenant (com token)
```bash
curl http://localhost:8081/api/v1/27PCKWWWN3F/config \
  -H "Authorization: Bearer <token>"
```

## � Segurança e URL Code

### Geração Automática de URL Code

O sistema gera automaticamente códigos de 11 caracteres aleatórios para isolamento seguro entre tenants.

**Características:**
- **Formato**: 11 caracteres (ex: `27PCKWWWN3F`)
- **Charset**: A-Z (uppercase) + 0-9 (36 possibilidades por char)
- **Entropia**: ~57 bits (36^11 = ~4 quintilhões de combinações)
- **Gerador**: `crypto/rand` (cryptographically secure)
- **Unicidade**: Verifica colisões no banco (retry até 10x)

**Implementação:**
```go
// internal/utils/code_generator.go
func GenerateURLCode() string {
    // Gera código seguro usando crypto/rand
    // Retorna: "27PCKWWWN3F"
}
```

**Por que não usar subdomain para admin?**
- 🔒 Segurança: Subdomain é público, url_code é privado
- 🎯 SEO: Subdomain é marketing, url_code é admin interno
- 🔐 Isolamento: Previne ataques de enumeração de tenants
- 🚀 Flexibilidade: Tenant pode mudar subdomain sem afetar admin

## �📁 Estrutura do Projeto

```
.
├── cmd/
│   ├── admin-api/        # Admin API (porta 8080)
│   ├── tenant-api/       # Tenant API (porta 8081)
│   └── worker/           # Worker de provisionamento
├── internal/
│   ├── cache/            # Cliente Redis
│   ├── config/           # Configurações
│   ├── database/         # Gerenciador de pools
│   ├── handlers/
│   │   ├── admin/        # Handlers do Control Plane
│   │   └── tenant/       # Handlers do Data Plane
│   ├── middleware/       # Middlewares (Auth, Tenant, Features)
│   ├── models/
│   │   ├── admin/        # Models para Control Plane
│   │   ├── tenant/       # DTOs para Data Plane
│   │   └── shared/       # Enums compartilhados
│   ├── repository/
│   │   └── admin/        # Acesso a dados do Master DB
│   ├── services/
│   │   └── admin/        # Lógica de negócio Control Plane
│   └── utils/            # Utilitários (JWT, hash, code generator)
├── migrations/
│   ├── master/           # Migrations Master DB
│   └── tenant/           # Migrations Tenant DB
├── config/
│   └── pgbouncer/        # Configuração PgBouncer
├── docs/
│   └── AUTH_FLOW.md      # Documentação detalhada do fluxo
├── scripts/              # Scripts utilitários
├── docker-compose.yml
├── Makefile
└── README.md
```

### 🏢 Separação por Domínio

**Control Plane (Admin)**: Gerenciamento de tenants, usuários, planos  
**Data Plane (Tenant)**: Operações dentro dos tenants isolados

```go
// Handlers organizados por domínio
internal/handlers/admin/    → users_handler.go, plans_handler.go
internal/handlers/tenant/   → auth_handler.go, products_handler.go

// Models separados por responsabilidade
internal/models/admin/      → Entidades do Master DB
internal/models/tenant/     → DTOs para comunicação
internal/models/shared/     → Enums compartilhados

// Repositórios focados
internal/repository/admin/  → Acesso exclusivo ao Master DB
// internal/repository/tenant/ → (futuro) Acesso aos bancos tenant
```

## 🔄 Provisionamento de Tenant

O sistema implementa provisionamento assíncrono automático:

1. **API** cria registro no Master DB (`status='provisioning'`)
2. **API** publica evento no Redis (`tenant:provision:{db_code}`)
3. **Worker** consome evento da fila Redis
4. **Worker** executa `CREATE DATABASE db_tenant_{db_code}`
5. **Worker** aplica migrations do Tenant DB
6. **Worker** atualiza status para `active`

**Tempo médio**: 2-5 segundos para provisionamento completo

### Verificar logs do Worker
```bash
make logs-worker
```

## 📝 Próximos Passos

- [x] Implementar Worker de provisionamento
- [x] Sistema de subscription público
- [x] Geração automática de url_code
- [x] Suporte a billing cycles
- [x] Reorganização da estrutura de diretórios por domínio
- [x] **Novo fluxo de autenticação com interface direta**
- [x] **Login retorna configuração completa do tenant**
- [x] **Endpoint de troca de tenant (switch-tenant)**
- [ ] Implementar CRUD completo de Produtos (Tenant DB)
- [ ] Implementar CRUD completo de Serviços (Tenant DB)
- [ ] Sistema de upload e gerenciamento de imagens
- [ ] Worker de processamento de imagens (resize, WebP)
- [ ] Configuração para múltiplos providers (Local/S3/R2)
- [ ] Admin API completa para gerenciamento de tenants
- [ ] Implementar rate limiting
- [ ] Adicionar logging estruturado
- [ ] Implementar métricas e observabilidade
- [ ] Sistema de pagamentos (Stripe/outros)
- [ ] Webhooks para eventos de tenant

## 🐛 Troubleshooting

### Tenant fica em `provisioning` para sempre
```bash
# Verificar logs do Worker
make logs-worker

# Verificar se o database foi criado
docker exec saas-postgres psql -U postgres -l | grep db_tenant
```

### Erro "subdomain already exists"
O subdomain escolhido já está em uso. Escolha outro nome único.

### Erro "url_code already exists" (raro)
Colisão de código aleatório. O sistema tenta 10x automaticamente. Se persistir, verifique o código.

### Worker não consome eventos
```bash
# Verificar se Redis está rodando
docker ps | grep redis

# Verificar fila no Redis
docker exec saas-redis redis-cli KEYS "tenant:provision:*"
```

### Reset completo do ambiente
```bash
make reset  # Remove volumes e recria tudo
```

## ❓ FAQ

**Q: Posso mudar o subdomain depois de criado?**  
A: Sim, mas requer update manual no banco. Planeje adicionar endpoint admin para isso.

**Q: url_code pode ser customizado?**  
A: Não diretamente via subscription. Apenas via Admin API (se implementado).

**Q: Quantos tenants o sistema suporta?**  
A: Limitado por PostgreSQL (teoricamente milhares). Use monitoring para escalar.

**Q: Como funciona o billing?**  
A: Atualmente apenas registra o `billing_cycle`. Integração com gateway de pagamento é próximo passo.

**Q: Posso ter múltiplos owners por tenant?**  
A: Não. Tenant tem um owner_id. Outros usuários são members com role específica.

## 📄 Licença

Este projeto é open source e está disponível sob a licença MIT.

---

**Desenvolvido com** ❤️ **usando Go 1.23**

