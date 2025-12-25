# WAYSCloud Terraform Provider

The official Terraform provider for [WAYSCloud](https://wayscloud.services) – a Nordic cloud platform providing IaaS, PaaS, and managed services on European infrastructure.

## About WAYSCloud

WAYSCloud is a Nordic cloud provider focused on **data sovereignty**, **open standards**, and **no vendor lock-in**. All infrastructure runs on European soil with full GDPR compliance.

- **Website:** https://wayscloud.services
- **Documentation:** https://docs.wayscloud.services
- **Security contact:** security@wayscloud.no

## Features

| Service | Resource | Description |
|---------|----------|-------------|
| **DNS** | `wayscloud_dns_zone` | Authoritative DNS hosting |
| **DNS** | `wayscloud_dns_record` | A, AAAA, CNAME, MX, TXT, SRV records |
| **Compute** | `wayscloud_vps` | Virtual Private Servers |
| **Storage** | `wayscloud_s3_bucket` | S3-compatible object storage |
| **Database** | `wayscloud_database` | Managed PostgreSQL & MariaDB |
| **Cache** | `wayscloud_redis_instance` | Managed Redis |
| **Apps** | `wayscloud_app` | Container platform with scale-to-zero |

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- WAYSCloud account with API key

## Installation

```hcl
terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloudas/wayscloud"
      version = "~> 0.1"
    }
  }
}

provider "wayscloud" {}
```

## Authentication

### API Key (Recommended)

```bash
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
```

Or in provider configuration:

```hcl
provider "wayscloud" {
  api_key = var.wayscloud_api_key  # Never hardcode!
}
```

### Personal Access Token (PAT)

Required for `wayscloud_database`. The provider auto-detects token type.

```bash
export WAYSCLOUD_API_KEY="wayscloud_pat_xxx..."
```

| Resource | Auth Type | Scopes |
|----------|-----------|--------|
| `wayscloud_dns_zone` | API Key | `dns` |
| `wayscloud_dns_record` | API Key | `dns` |
| `wayscloud_redis_instance` | API Key | `redis` |
| `wayscloud_s3_bucket` | API Key | `storage` |
| `wayscloud_vps` | API Key | `vps` |
| `wayscloud_app` | API Key | `apps` |
| `wayscloud_database` | **PAT** | `database:read`, `database:write` |

## Quick Start

```hcl
# Create a DNS zone and record
resource "wayscloud_dns_zone" "example" {
  name = "example.com"
}

resource "wayscloud_dns_record" "www" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}

# Create a Redis cache
resource "wayscloud_redis_instance" "cache" {
  name   = "my-cache"
  region = "no"
  plan   = "redis-starter"
}
```

## Import Existing Resources

```bash
# DNS Zone
terraform import wayscloud_dns_zone.example example.com

# DNS Record
terraform import wayscloud_dns_record.www example.com/RECORD_UUID

# Redis Instance
terraform import wayscloud_redis_instance.cache INSTANCE_UUID

# S3 Bucket
terraform import wayscloud_s3_bucket.uploads bucket-name

# Database
terraform import wayscloud_database.app postgresql/db-name

# VPS
terraform import wayscloud_vps.web VPS_UUID

# App
terraform import wayscloud_app.api app_ULID
```

## Known Limitations

### Async Resources

Some resources have async provisioning. The provider polls until ready:

| Resource | Typical Wait Time |
|----------|-------------------|
| `wayscloud_redis_instance` | 2-5 minutes |
| `wayscloud_vps` | 3-10 minutes |

### Sensitive Attributes

These attributes are only available on initial creation:

- `wayscloud_database.password`, `wayscloud_database.connection_string`
- `wayscloud_redis_instance.password`
- `wayscloud_s3_bucket.secret_key`

## Rate Limits & Retries

The provider automatically:
- Retries transient errors (429, 502, 503, 504) up to 3 times
- Uses exponential backoff between retries
- Times out requests after 30 seconds

## Development

```bash
# Build
go build -o terraform-provider-wayscloud

# Test
go test ./... -race

# Acceptance tests
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
export TF_ACC=1
go test ./... -v -timeout 60m
```

## Versioning

This provider follows [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes to resource schemas
- **MINOR**: New resources or attributes
- **PATCH**: Bug fixes and documentation

## Security

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

**Never commit API keys or tokens to version control.**

## License

[Mozilla Public License 2.0](LICENSE)

## Support

- [Terraform Registry Docs](https://registry.terraform.io/providers/wayscloudas/wayscloud/latest/docs)
- [GitHub Issues](https://github.com/wayscloudas/terraform-provider-wayscloud/issues)
- [WAYSCloud Support](https://wayscloud.services/support)
