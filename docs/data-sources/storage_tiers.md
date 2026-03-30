---
page_title: "wayscloud_storage_tiers Data Source - terraform-provider-wayscloud"
subcategory: ""
description: |-
  Fetches the list of available WAYSCloud S3 storage tiers.
---

# wayscloud_storage_tiers (Data Source)

Fetches the list of available WAYSCloud S3 storage tiers.

## Example Usage

```hcl
data "wayscloud_storage_tiers" "all" {}

output "available_tiers" {
  value = data.wayscloud_storage_tiers.all.tiers
}
```

## Attribute Reference

- `tiers` - List of available storage tiers. Each entry contains:
  - `id` - Tier identifier.
  - `name` - Tier display name.
  - `description` - Human-readable description of the storage tier.
  - `monthly_price` - Monthly price for the tier.
  - `currency` - Price currency (e.g., `NOK`, `EUR`).
