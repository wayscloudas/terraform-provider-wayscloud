---
page_title: "wayscloud_redis_instance Resource - WAYSCloud"
description: |-
  Manages a Redis instance in WAYSCloud.
---

# wayscloud_redis_instance (Resource)

Manages a Redis instance in WAYSCloud. Redis instances are provisioned asynchronously and the provider will poll until the instance is ready.

## Example Usage

```terraform
resource "wayscloud_redis_instance" "cache" {
  name   = "my-app-cache"
  region = "no"
  plan   = "redis-starter"
}

output "redis_host" {
  value = wayscloud_redis_instance.cache.host
}
```

## Schema

### Required

- `name` (String) - Instance name. Changing this forces a new resource.
- `region` (String) - Region code (e.g., "no"). Changing this forces a new resource.
- `plan` (String) - Plan code (e.g., "redis-starter", "redis-pro"). Changing this forces a new resource.

### Read-Only

- `id` (String) - The instance UUID.
- `host` (String) - Redis host address.
- `port` (Number) - Redis port.
- `password` (String, Sensitive) - Redis password. Only available on creation.
- `status` (String) - Instance status.
- `created_at` (String) - Timestamp when created.

## Import

Redis instances can be imported using the UUID:

```bash
terraform import wayscloud_redis_instance.cache 550e8400-e29b-41d4-a716-446655440000
```

~> **Note:** The `password` attribute will not be available after import.
