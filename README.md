# Multi-Tenant SaaS API

API SaaS escalável em Go com isolamento de dados por banco físico (database-per-tenant).

## 🚀 Características

- **Database-per-Tenant**: Cada tenant possui seu próprio banco de dados físico para completo isolamento de dados
- **Control Plane**: Banco Master centralizado para gerenciamento de usuários, tenants, planos e RBAC
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
make test-admin-login    # Testar login Admin API
make test-tenant-login   # Testar login Tenant API
make test-tenant         # Criar tenant via Admin API
make test-subscription   # Testar cadastro público

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
  "user": {
    "id": "e95b2979-e1c6-4ded-8a36-3340f78ff931",
    "email": "joao@teste.com"
  },
  "tenant": {
    "id": "057d0d5c-415f-4bc2-a8fb-2a9bd524076d",
    "url_code": "27PCKWWWN3F",
    "subdomain": "joao",
    "billing_cycle": "monthly",
    "status": "provisioning"
  }
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
    "email": "joao@teste.com"
  },
  "tenants": [
    {
      "id": "tenant-uuid",
      "url_code": "27PCKWWWN3F",
      "subdomain": "joao",
      "name": "João Silva",
      "role": "owner"
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

### 📊 Resumo de Endpoints

| Método | Endpoint | Auth | Descrição |
|--------|----------|------|-----------|
| `POST` | `/api/v1/subscription` | ❌ Público | Cadastro de novo assinante |
| `POST` | `/api/v1/auth/login` | ❌ Público | Login tenant |
| `GET` | `/api/v1/auth/me` | ✅ JWT | Dados do usuário logado |
| `GET` | `/api/v1/adm/:url_code/config` | ✅ JWT + Tenant | Config do frontend |
| `GET` | `/api/v1/adm/:url_code/products` | ✅ JWT + Feature | Listar produtos |
| `POST` | `/api/v1/adm/:url_code/products` | ✅ JWT + Permission | Criar produto |
| `GET` | `/api/v1/adm/:url_code/services` | ✅ JWT + Feature | Listar serviços |
| `POST` | `/api/v1/adm/:url_code/services` | ✅ JWT + Permission | Criar serviço |
| `POST` | `/api/v1/admin/login` | ❌ Público | Login admin (porta 8080) |
| `POST` | `/api/v1/admin/tenants` | ✅ Admin JWT | Criar tenant (admin) |

**Legenda:**
- ✅ JWT: Requer header `Authorization: Bearer <token>`
- ✅ JWT + Tenant: Requer acesso ao tenant via `tenant_members`
- ✅ JWT + Feature: Requer feature habilitada no plano
- ✅ JWT + Permission: Requer permissão específica do usuário

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

### 3. Fazer login
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "joao@teste.com",
    "password": "senha12345"
  }'
```

### 4. Acessar admin panel (com token)
```bash
curl http://localhost:8081/api/v1/adm/27PCKWWWN3F/config \
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
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # Middlewares (Auth, Tenant)
│   ├── models/           # Modelos de dados
│   ├── repository/       # Camada de acesso a dados
│   ├── services/         # Lógica de negócio
│   └── utils/            # Utilitários (JWT, hash, code generator)
├── migrations/
│   ├── master/           # Migrations Master DB
│   └── tenant/           # Migrations Tenant DB
├── config/
│   └── pgbouncer/        # Configuração PgBouncer
├── docker-compose.yml
├── Makefile
└── README.md
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
- [ ] Implementar Admin API completa
- [ ] Adicionar endpoints de gerenciamento de tenants
- [ ] Implementar handlers completos de Products/Services
- [ ] Adicionar testes unitários e de integração
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

