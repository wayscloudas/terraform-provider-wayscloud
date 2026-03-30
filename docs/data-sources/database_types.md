---
page_title: "wayscloud_database_types Data Source - terraform-provider-wayscloud"
subcategory: ""
description: |-
  Fetches the list of available WAYSCloud database types and tiers.
---

# wayscloud_database_types (Data Source)

Fetches the list of available WAYSCloud database types and tiers.

## Example Usage

```hcl
data "wayscloud_database_types" "all" {}

output "available_types" {
  value = data.wayscloud_database_types.all.database_types
}
```

## Attribute Reference

- `database_types` - List of available database types. Each entry contains:
  - `type` - Database engine type (e.g., `postgresql`, `mariadb`).
  - `tier` - Tier level (e.g., `standard`, `encrypted`).
  - `description` - Human-readable description of the database type.
