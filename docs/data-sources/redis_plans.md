---
page_title: "wayscloud_redis_plans Data Source - terraform-provider-wayscloud"
subcategory: ""
description: |-
  Fetches the list of available WAYSCloud Redis plans.
---

# wayscloud_redis_plans (Data Source)

Fetches the list of available WAYSCloud Redis plans.

## Example Usage

```hcl
data "wayscloud_redis_plans" "all" {}

output "available_plans" {
  value = data.wayscloud_redis_plans.all.plans
}
```

## Attribute Reference

- `plans` - List of available Redis plans. Each entry contains:
  - `id` - Plan identifier.
  - `name` - Plan display name.
  - `memory_mb` - Memory allocation in megabytes.
  - `monthly_price` - Monthly price for the plan.
  - `currency` - Price currency (e.g., `NOK`, `EUR`).
