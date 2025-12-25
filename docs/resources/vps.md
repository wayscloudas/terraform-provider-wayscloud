---
page_title: "wayscloud_vps Resource - WAYSCloud"
description: |-
  Manages a Virtual Private Server (VPS) in WAYSCloud.
---

# wayscloud_vps (Resource)

Manages a Virtual Private Server (VPS) in WAYSCloud. VPS instances are provisioned asynchronously and the provider will poll until the server is ready.

## Example Usage

```terraform
resource "wayscloud_vps" "web" {
  hostname    = "web-server-01"
  plan_code   = "vps-starter"
  region      = "no"
  os_template = "ubuntu-22.04"
  ssh_keys    = ["ssh-ed25519 AAAA... user@host"]
}

output "ip_address" {
  value = wayscloud_vps.web.ip_address
}
```

## Schema

### Required

- `hostname` (String) - Server hostname. Changing this forces a new resource.
- `plan_code` (String) - Plan code. Changing this forces a new resource.
- `region` (String) - Region code. Changing this forces a new resource.
- `os_template` (String) - OS template. Changing this forces a new resource.

### Optional

- `ssh_keys` (List of String) - SSH public keys for root access. Changing this forces a new resource.

### Read-Only

- `id` (String) - The VPS UUID.
- `ip_address` (String) - Primary IPv4 address.
- `ip_address_v6` (String) - Primary IPv6 address.
- `status` (String) - Server status.
- `created_at` (String) - Timestamp when created.

## Import

VPS instances can be imported using the UUID:

```bash
terraform import wayscloud_vps.web 550e8400-e29b-41d4-a716-446655440000
```
