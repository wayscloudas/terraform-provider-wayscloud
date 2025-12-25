---
page_title: "wayscloud_database Resource - WAYSCloud"
description: |-
  Manages a PostgreSQL or MariaDB database in WAYSCloud.
---

# wayscloud_database (Resource)

Manages a PostgreSQL or MariaDB database in WAYSCloud.

~> **Important:** This resource requires a Personal Access Token (PAT) with `database:read` and `database:write` scopes. API keys are not supported.

## Example Usage

```terraform
resource "wayscloud_database" "app" {
  name = "myapp-prod"
  type = "postgresql"
  tier = "standard"
}

output "connection_string" {
  value     = wayscloud_database.app.connection_string
  sensitive = true
}
```

## Schema

### Required

- `name` (String) - Database name. Changing this forces a new resource.
- `type` (String) - Database type: "postgresql" or "mariadb". Changing this forces a new resource.

### Optional

- `tier` (String) - Storage tier: "standard" or "encrypted". Default: "standard". Changing this forces a new resource.

### Read-Only

- `id` (String) - The database UUID.
- `host` (String) - Database host address.
- `port` (Number) - Database port.
- `username` (String) - Database username.
- `password` (String, Sensitive) - Database password. Only available on creation.
- `connection_string` (String, Sensitive) - Full connection string. Only available on creation.
- `status` (String) - Database status.
- `created_at` (String) - Timestamp when created.

## Import

Databases can be imported using type/name:

```bash
terraform import wayscloud_database.app postgresql/myapp-prod
```

~> **Note:** The `password` and `connection_string` attributes will not be available after import.
