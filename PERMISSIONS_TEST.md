# Test Login Permissions Script

## Overview
This script tests if permissions are correctly included in the login payload for both owners and non-owner users.

## Test Cases

### 1. Owner User Login
- **Expected**: All permissions returned in payload
- **Role**: "owner" 
- **Permissions**: All available permissions (from permissions table)

### 2. Non-Owner User Login  
- **Expected**: Only role-specific permissions returned
- **Role**: Based on tenant_members role assignment
- **Permissions**: Only permissions assigned to the user's role

## Repository Changes Made

### GetUserPermissions()
- **Before**: Only queried permissions via role for ALL users
- **After**: 
  - Check if user is tenant owner (owner_id match)
  - If owner: return ALL permissions 
  - If non-owner: return role-based permissions

### GetUserRole()
- **Before**: Only queried role from tenant_members
- **After**:
  - Check if user is tenant owner (owner_id match) 
  - If owner: return "owner"
  - If non-owner: return role from tenant_members

### GetUserTenants()
- **Before**: Only found tenants via tenant_members table
- **After**: 
  - Include tenants where user is owner_id
  - Use CASE statement to set role as "owner" for owned tenants
  - Use LEFT JOIN to include both member and owned tenants

## Expected Login Response

### Owner Login
```json
{
  "token": "...",
  "user": {...},
  "tenants": [...],
  "current_tenant": {...},
  "features": ["products", "services"],
  "permissions": [
    "create_product", "read_product", "update_product", "delete_product",
    "create_service", "read_service", "update_service", "delete_service", 
    "manage_users", "manage_settings", "setg_m", "prod_c", "prod_r", 
    "prod_u", "prod_d", "serv_c", "serv_r", "serv_u", "serv_d"
  ]
}
```

### Non-Owner Login (example: user with limited role)
```json
{
  "token": "...",
  "user": {...}, 
  "tenants": [...],
  "current_tenant": {...},
  "features": ["products", "services"],
  "permissions": [
    "read_product", "read_service"
  ]
}
```

## Testing

1. **Database Setup**: Ensure permissions table has sample permissions
2. **Create Owner**: User who is tenant.owner_id  
3. **Create Non-Owner**: User in tenant_members with limited role
4. **Login Both**: Compare permission arrays in responses
5. **Middleware Test**: Verify permissions injected into context correctly

## Key Benefit
- Owners now automatically get all permissions without explicit role assignment
- Non-owners get permissions based on their assigned roles
- Login payload correctly reflects user capabilities for frontend permission checks