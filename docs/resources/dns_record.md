---
page_title: "wayscloud_dns_record Resource - WAYSCloud"
description: |-
  Manages a DNS record in WAYSCloud.
---

# wayscloud_dns_record (Resource)

Manages a DNS record in a WAYSCloud DNS zone.

## Example Usage

```terraform
resource "wayscloud_dns_record" "www" {
  zone_name = "example.com"
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}

resource "wayscloud_dns_record" "mail" {
  zone_name = "example.com"
  name      = "@"
  type      = "MX"
  value     = "mail.example.com"
  priority  = 10
  ttl       = 3600
}
```

## Schema

### Required

- `zone_name` (String) - The DNS zone name. Changing this forces a new resource.
- `name` (String) - Record name (e.g., "www", "@" for apex). Changing this forces a new resource.
- `type` (String) - Record type: A, AAAA, CNAME, MX, TXT, SRV, NS, PTR. Changing this forces a new resource.
- `value` (String) - Record value.
- `ttl` (Number) - Time to live in seconds.

### Optional

- `priority` (Number) - Priority for MX and SRV records.

### Read-Only

- `id` (String) - The record UUID.

## Import

DNS records can be imported using zone_name/record_id:

```bash
terraform import wayscloud_dns_record.www example.com/550e8400-e29b-41d4-a716-446655440000
```
