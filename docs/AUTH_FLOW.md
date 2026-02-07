# Fluxo de Autenticação Multi-Tenant

## Visão Geral

O sistema utiliza um fluxo de autenticação simplificado onde o usuário faz **login uma única vez** e recebe automaticamente a configuração do último tenant acessado (`last_tenant_id`).

## Endpoints

### 1. Login do Usuário

**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "email": "usuario@exemplo.com",
  "password": "senha123"
}
```

**Response:**
```json
{
  "token": "JWT_TOKEN",
  "user": {
    "id": "user-uuid",
    "email": "usuario@exemplo.com",
    "full_name": "Nome do Usuário"
  },
  "tenants": [
    {
      "id": "tenant-uuid-1",
      "url_code": "empresa-x",
      "subdomain": "empresax",
      "name": "Empresa X",
      "role": "owner"
    },
    {
      "id": "tenant-uuid-2",
      "url_code": "empresa-y",
      "subdomain": "empresay",
      "name": "Empresa Y",
      "role": "member"
    }
  ],
  "current_tenant": {
    "id": "tenant-uuid-1",
    "url_code": "empresa-x",
    "subdomain": "empresax",
    "name": "Empresa X"
  },
  "interface": {
    "logo_url": "https://cdn.example.com/logo.png",
    "company_name": "Empresa X Ltda",
    "custom_settings": {
      "primary_color": "#3B82F6",
      "secondary_color": "#10B981",
      "theme": "light"
    }
  },
  "features": ["products", "services", "customers"],
  "permissions": ["prod_r", "prod_c", "prod_u", "serv_r", "serv_c"]
}
```

**Comportamento:**
- Se o usuário tem `last_tenant_logged` configurado, o sistema retorna:
  - `current_tenant`: Dados do último tenant acessado
  - `interface`: Configuração de layout do tenant (logo, cores, etc)
  - `features`: Features disponíveis no plano do tenant
  - `permissions`: Permissões do usuário naquele tenant
- Se o usuário nunca acessou nenhum tenant (novo usuário), esses campos ficam `null`

### 2. Trocar de Tenant

**Endpoint:** `POST /api/v1/auth/switch-tenant`

**Headers:**
```
Authorization: Bearer JWT_TOKEN
```

**Request:**
```json
{
  "url_code": "empresa-y"
}
```

**Response:**
```json
{
  "message": "tenant switched successfully",
  "current_tenant": {
    "id": "tenant-uuid-2",
    "url_code": "empresa-y",
    "subdomain": "empresay",
    "name": "Empresa Y"
  },
  "interface": {
    "logo_url": "https://cdn.example.com/logo-y.png",
    "company_name": "Empresa Y S/A",
    "custom_settings": {
      "primary_color": "#EF4444",
      "secondary_color": "#F59E0B",
      "theme": "dark"
    }
  },
  "features": ["products", "customers"],
  "permissions": ["prod_r", "cust_r", "cust_c"]
}
```

**Comportamento:**
- Valida que o usuário tem acesso ao tenant solicitado
- Atualiza o campo `last_tenant_logged` no banco de dados
- Retorna a nova configuração de interface, features e permissões

## Fluxo no Frontend

### Login Inicial

```typescript
// 1. Usuário faz login
const loginResponse = await api.post('/auth/login', {
  email: 'usuario@exemplo.com',
  password: 'senha123'
});

// 2. Armazena token
localStorage.setItem('token', loginResponse.data.token);
localStorage.setItem('user', JSON.stringify(loginResponse.data.user));
localStorage.setItem('tenants', JSON.stringify(loginResponse.data.tenants));

// 3. Se tem current_tenant, configura a interface
if (loginResponse.data.current_tenant) {
  localStorage.setItem('current_tenant', JSON.stringify(loginResponse.data.current_tenant));
  localStorage.setItem('interface', JSON.stringify(loginResponse.data.interface));
  localStorage.setItem('features', JSON.stringify(loginResponse.data.features));
  localStorage.setItem('permissions', JSON.stringify(loginResponse.data.permissions));
  
  // Redireciona para dashboard
  router.push('/dashboard');
} else {
  // Usuário novo sem tenants - mostrar tela de criação de tenant
  router.push('/create-tenant');
}
```

### Troca de Tenant

```typescript
// 1. Usuário seleciona outro tenant do menu
const switchResponse = await api.post('/auth/switch-tenant', {
  url_code: 'empresa-y'
}, {
  headers: { Authorization: `Bearer ${token}` }
});

// 2. Atualiza configurações locais
localStorage.setItem('current_tenant', JSON.stringify(switchResponse.data.current_tenant));
localStorage.setItem('interface', JSON.stringify(switchResponse.data.interface));
localStorage.setItem('features', JSON.stringify(switchResponse.data.features));
localStorage.setItem('permissions', JSON.stringify(switchResponse.data.permissions));

// 3. Recarrega a interface ou redireciona
window.location.reload(); // ou usar state management (Vuex/Pinia/Redux)
```

### Uso da Interface Config

```typescript
// Em qualquer componente Vue/React
const interface = JSON.parse(localStorage.getItem('interface') || '{}');

// Template
<div :style="{
  '--primary-color': interface.custom_settings?.primary_color,
  '--secondary-color': interface.custom_settings?.secondary_color
}">
  <img :src="interface.logo_url" alt="Logo" />
  <h1>{{ interface.company_name }}</h1>
</div>
```

### Verificação de Features

```typescript
const features = JSON.parse(localStorage.getItem('features') || '[]');

// Mostrar menu apenas se feature estiver disponível
const showProductsMenu = features.includes('products');
const showServicesMenu = features.includes('services');
```

### Verificação de Permissões

```typescript
const permissions = JSON.parse(localStorage.getItem('permissions') || '[]');

// Habilitar botões com base em permissões
const canCreateProduct = permissions.includes('prod_c');
const canDeleteProduct = permissions.includes('prod_d');
```

## Vantagens deste Fluxo

1. **Single Sign-On**: Usuário faz login uma vez e já entra direto no sistema
2. **Persistência**: Sistema lembra do último tenant acessado
3. **Performance**: Todas as informações necessárias vêm em uma única chamada
4. **UX Simplificada**: Sem necessidade de selecionar tenant a cada login
5. **Multi-Tenant Suave**: Troca de tenant é rápida e simples
6. **Configuração Dinâmica**: Interface se adapta automaticamente ao tenant ativo

## Casos de Uso

### Usuário com 1 Tenant
- Login → Entra direto no dashboard com configurações carregadas
- Experiência fluida, sem passos extras

### Usuário com Múltiplos Tenants
- Login → Entra no último tenant usado
- Menu dropdown mostra lista de tenants disponíveis
- Usuário pode trocar de tenant a qualquer momento via switch-tenant
- Sistema atualiza interface/features/permissions automaticamente

### Novo Usuário (sem tenants)
- Login → Recebe token e lista vazia de tenants
- Frontend detecta ausência de `current_tenant`
- Redireciona para tela de criação de tenant ou convite

## Segurança

- **JWT**: Token contém apenas `user_id` (não contém tenant_id para permitir multi-tenant)
- **Validação de Acesso**: Switch-tenant valida que usuário tem acesso ao tenant solicitado
- **Features**: Validadas no backend antes de executar operações
- **Permissions**: Verificadas em cada endpoint que modifica dados
- **CORS**: Configurado para aceitar apenas domínios autorizados
- **Rate Limiting**: Previne abuso dos endpoints de autenticação

## Migrações Futuras

Se precisar dar suporte ao fluxo antigo (dois passos), pode-se manter ambos:
- `/auth/login` - Novo fluxo (recomendado)
- `/auth/login-to-tenant` - Fluxo legado (deprecated)

Mas como o código foi refatorado, o recomendado é usar **apenas o novo fluxo**.
