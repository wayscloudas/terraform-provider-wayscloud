---
page_title: "wayscloud_dns_zone Resource - WAYSCloud"
description: |-
  Manages a DNS zone in WAYSCloud.
---

# wayscloud_dns_zone (Resource)

Manages a DNS zone in WAYSCloud.

## Example Usage

```terraform
resource "wayscloud_dns_zone" "example" {
  name = "example.com"
}
```

## Schema

### Required

- `name` (String) - The domain name for the DNS zone (e.g., "example.com"). Changing this forces a new resource.

### Read-Only

- `id` (String) - The zone identifier (same as name).
- `created_at` (String) - Timestamp when the zone was created.

## Import

DNS zones can be imported using the domain name:

```bash
terraform import wayscloud_dns_zone.example example.com
```
