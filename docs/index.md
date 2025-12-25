---
page_title: "WAYSCloud Provider"
description: |-
  The WAYSCloud provider allows you to manage WAYSCloud infrastructure resources.
---

# WAYSCloud Provider

The WAYSCloud provider allows you to manage [WAYSCloud](https://wayscloud.services) infrastructure using Terraform.

## Example Usage

```terraform
terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloudas/wayscloud"
      version = "~> 0.1.0"
    }
  }
}

provider "wayscloud" {
  # API key from environment variable WAYSCLOUD_API_KEY
}

# Create a DNS zone
resource "wayscloud_dns_zone" "example" {
  name = "example.com"
}

# Create an A record
resource "wayscloud_dns_record" "www" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}
```

## Authentication

The provider supports two authentication methods:

### API Key (Recommended)

```bash
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
```

Or in provider configuration:

```terraform
provider "wayscloud" {
  api_key = var.wayscloud_api_key
}
```

### Personal Access Token (PAT)

Required for `wayscloud_database` resource:

```bash
export WAYSCLOUD_API_KEY="wayscloud_pat_xxx..."
```

The provider auto-detects the token type based on the prefix.

## Resources

| Resource | Description |
|----------|-------------|
| [wayscloud_dns_zone](resources/dns_zone) | Manage DNS zones |
| [wayscloud_dns_record](resources/dns_record) | Manage DNS records |
| [wayscloud_redis_instance](resources/redis_instance) | Managed Redis instances |
| [wayscloud_s3_bucket](resources/s3_bucket) | S3-compatible storage buckets |
| [wayscloud_database](resources/database) | Managed PostgreSQL/MariaDB |
| [wayscloud_vps](resources/vps) | Virtual Private Servers |
| [wayscloud_app](resources/app) | Container app platform |

## Schema

### Optional

- `api_key` (String, Sensitive) - WAYSCloud API key or PAT token. Can also be set via `WAYSCLOUD_API_KEY` environment variable.
- `endpoint` (String) - API endpoint URL. Defaults to `https://api.wayscloud.services`. Can also be set via `WAYSCLOUD_ENDPOINT` environment variable.
