# WAYSCloud Terraform Provider

The official Terraform provider for [WAYSCloud](https://wayscloud.services) – a Nordic cloud platform providing IaaS, PaaS, and managed services on European infrastructure.

## About WAYSCloud

WAYSCloud is a Nordic cloud provider focused on **data sovereignty**, **open standards**, and **no vendor lock-in**. All infrastructure runs on European soil with full GDPR compliance.

- **Website:** https://wayscloud.services
- **Documentation:** https://docs.wayscloud.services
- **Security contact:** security@wayscloud.net

## Features

| Service | Resource | Description |
|---------|----------|-------------|
| **DNS** | `wayscloud_dns_zone` | Authoritative DNS hosting |
| **DNS** | `wayscloud_dns_record` | A, AAAA, CNAME, MX, TXT, SRV records |
| **Compute** | `wayscloud_vps` | Virtual Private Servers (Linux & Windows) |
| **Storage** | `wayscloud_s3_bucket` | S3-compatible object storage |
| **Database** | `wayscloud_database` | Managed PostgreSQL & MariaDB |
| **Cache** | `wayscloud_redis_instance` | Managed Redis |
| **Apps** | `wayscloud_app` | Container platform with scale-to-zero |
| **IoT** | `wayscloud_iot_device` | IoT device management with MQTT |
| **IoT** | `wayscloud_iot_device_group` | IoT device grouping |
| **IoT** | `wayscloud_iot_rule` | IoT rule management |
| **Storage** | `wayscloud_s3_bucket_key` | S3 access keys |
| **Domains** | `wayscloud_domain_verification` | Domain ownership verification |

**Data sources:** `wayscloud_regions`, `wayscloud_dns_zones`, `wayscloud_vps_plans`, `wayscloud_vps_os_templates`, `wayscloud_app_plans`, `wayscloud_database_types`, `wayscloud_redis_plans`, `wayscloud_storage_tiers`

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (for building from source)
- WAYSCloud account with API key and/or PAT token

## Installation

```hcl
terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloudas/wayscloud"
      version = "~> 0.4"
    }
  }
}

provider "wayscloud" {}
```

## Authentication

The provider supports two authentication methods that can be used independently or together:

| Auth Type | Header | Used For | Environment Variable |
|-----------|--------|----------|---------------------|
| **API Key** | `X-API-Key` | DNS, VPS, Storage, Redis, IoT, SMS, Apps | `WAYSCLOUD_API_KEY` |
| **PAT Token** | `Authorization: Bearer` | Database, Domain Verification | `WAYSCLOUD_PAT_TOKEN` |

### Option 1: Environment variables (recommended)

```bash
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."       # For DNS, VPS, Storage, etc.
export WAYSCLOUD_PAT_TOKEN="wayscloud_pat_xxx..."     # For Database, Domain Verification
```

### Option 2: Provider configuration

```hcl
provider "wayscloud" {
  api_key   = var.wayscloud_api_key    # For DNS, VPS, Storage, Redis, IoT, SMS, Apps
  pat_token = var.wayscloud_pat_token  # For Database, Domain Verification
}
```

### Auth requirements per resource

| Resource | Auth Type | Scopes |
|----------|-----------|--------|
| `wayscloud_dns_zone` | API Key | `dns` |
| `wayscloud_dns_record` | API Key | `dns` |
| `wayscloud_vps` | API Key | `vps` |
| `wayscloud_s3_bucket` | API Key | `storage` |
| `wayscloud_redis_instance` | API Key | `redis` |
| `wayscloud_app` | API Key | `apps` |
| `wayscloud_iot_device` | API Key | `iot` |
| `wayscloud_database` | **PAT** | `database:read`, `database:write` |
| `wayscloud_domain_verification` | **PAT** | `domain-verification` |

### Getting credentials

1. **API Key:** Log in to [my.wayscloud.services](https://my.wayscloud.services) → Account → API Keys → Create API Key
2. **PAT Token:** Log in to [my.wayscloud.services](https://my.wayscloud.services) → Account → Personal Access Tokens → Create Token (select required scopes)

## Quick Start

```hcl
# DNS zone and record
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

# VPS with Ubuntu
resource "wayscloud_vps" "web" {
  hostname    = "web01.example.com"
  plan_code   = "NO-Start-Linux-2cpu-4096mb-30gb"
  region      = "NO"
  os_template = "ubuntu-24.04"

  ssh_keys = ["ssh-rsa AAAAB3..."]
}

# Managed database (requires PAT)
resource "wayscloud_database" "app" {
  name = "myapp-prod"
  type = "postgresql"
  tier = "standard"
}
```

## Import Existing Resources

```bash
# DNS Zone
terraform import wayscloud_dns_zone.example example.com

# DNS Record
terraform import wayscloud_dns_record.www example.com/RECORD_UUID

# VPS
terraform import wayscloud_vps.web VPS_UUID

# S3 Bucket
terraform import wayscloud_s3_bucket.uploads bucket-name

# Redis Instance
terraform import wayscloud_redis_instance.cache INSTANCE_UUID

# Database (format: type/tier/name)
terraform import wayscloud_database.app postgresql/standard/db-name

# App
terraform import wayscloud_app.api app_ULID

# IoT Device
terraform import wayscloud_iot_device.sensor temp-sensor-01

# Domain Verification
terraform import wayscloud_domain_verification.email VERIFICATION_UUID
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
- `wayscloud_iot_device.mqtt_username`, `wayscloud_iot_device.mqtt_password`

## Rate Limits & Retries

The provider automatically:
- Retries transient errors (429, 502, 503, 504) up to 3 times
- Uses exponential backoff between retries
- Times out requests after 30 seconds

## Development

```bash
# Build
make build

# Test
make test

# Acceptance tests
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
export WAYSCLOUD_PAT_TOKEN="wayscloud_pat_xxx..."
make testacc

# Install locally
make install
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

- [Documentation](https://docs.wayscloud.services/integrations/terraform)
- [Terraform Registry](https://registry.terraform.io/providers/wayscloudas/wayscloud/latest/docs)
- [GitHub Issues](https://github.com/wayscloudas/terraform-provider-wayscloud/issues)
- [WAYSCloud Support](https://my.wayscloud.services/support)
