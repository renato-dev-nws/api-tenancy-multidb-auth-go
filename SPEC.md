# API Go Multi-Tenant (Database-per-Tenant)

## 1. Visão Geral

API SaaS escalável em **Go**, com isolamento de dados por banco físico (`db_tenant_{uuid}`). O sistema utiliza um **Control Plane** central para gerenciar assinaturas, planos baseados em features granulares e controle de acesso (RBAC).

## 2. Stack Tecnológica

* **Linguagem:** Go 1.21+ (Framework: Gin ou Echo)
* **Banco de Dados:** PostgreSQL (Driver: `pgx/v5` com `pgxpool`)
* **Pooler de Conexões:** PgBouncer (Modo: Transaction)
* **Cache & Message Broker:** Redis
* **Infraestrutura:** Docker & Docker Compose

## 3. Arquitetura de Dados (Master DB - Control Plane)

### 3.1. Gestão de Identidade e Tenants

* **Table `users**`: `id` (UUID), `email`, `password_hash`, `last_tenant_id`.
* **Table `user_profiles**`: `user_id`, `full_name`, `avatar_url`.
* **Table `tenants**`:
* `id`: UUID (PK).
* `db_code`: UUID (Nome físico: `db_tenant_{db_code}`).
* `url_code`: string (11 chars, Unique, ex: `2RT64H3JD77`).
* `owner_id`: FK users.
* `plan_id`: FK plans.
* `status`: enum (provisioning, active, suspended).


* **Table `tenant_profiles**`: `tenant_id`, `logo_url`, `custom_settings` (JSONB).
* **Table `tenant_members**`: `user_id`, `tenant_id`, `role_id` (vínculo de usuários a múltiplos tenants).

### 3.2. Planos e Features (Dynamic Modules)

* **Table `features**`: `id`, `title`, `slug` (ex: "products", "services"), `description`.
* **Table `plans**`: `id`, `name`, `description`, `price`.
* **Table `plan_features**`: `plan_id`, `feature_id` (Relacionamento Many-to-Many).

### 3.3. RBAC (Role Based Access Control)

* **Table `roles**`: `id`, `tenant_id` (ou NULL para roles globais), `name`, `slug`.
* **Table `permissions**`: `id`, `name`, `slug` (ex: "create_product", "delete_user").
* **Table `role_permissions**`: `role_id`, `permission_id`.

## 4. Arquitetura de Dados (Tenant DB - Individual)

* **Table `products**`, **`services`**, **`settings`**.
* *Nota: Cada banco é criado dinamicamente via Worker.*

## 5. Lógica de Negócio e Segurança

### 5.1. Roteamento e Resolução

1. O Middleware captura o `:url_code` da rota ou subdomínio.
2. Busca no cache (Redis) o `db_code` correspondente.
3. Verifica na tabela `tenant_members` se o usuário logado tem permissão de acesso.
4. Consulta as `features` vinculadas ao plano do tenant.
5. Injeta o pool de conexão (`*pgxpool.Pool`) e a lista de features ativas no `context.Context`.

### 5.2. Frontend Bridge (VueJS)

* Endpoint `GET /adm/:url_code/config` deve retornar:
* Lista de **features** (pelo slug) para montagem do menu dinâmico.
* Lista de **permissions** do usuário logado para habilitar/desabilitar botões na UI.



### 5.3. Worker de Provisionamento (Async)

1. **Trigger:** API cria registro no Master e dispara evento no Redis.
2. **Action:** Worker consome evento, executa `CREATE DATABASE "db_tenant_{db_code}"` no Postgres.
3. **Migration:** Worker aplica o schema inicial (products, services) no novo banco.

## 6. Configuração de Infraestrutura

* **PgBouncer:** Porta 6432, mapeando dinamicamente para os bancos `db_tenant_*`.
* **Segurança:** A API utiliza um usuário de banco com privilégios limitados (não-superuser).
