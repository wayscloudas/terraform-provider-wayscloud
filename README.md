# Terraform Provider for WAYSCloud

The WAYSCloud provider allows you to manage [WAYSCloud](https://wayscloud.services) infrastructure using Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloud/wayscloud"
      version = "~> 0.1.0"
    }
  }
}

provider "wayscloud" {}
```

### From Source

```bash
git clone https://github.com/wayscloud/terraform-provider-wayscloud.git
cd terraform-provider-wayscloud
go build -o terraform-provider-wayscloud
```

## Authentication

The provider supports two authentication methods:

### API Key (Recommended for most resources)

API keys authenticate with the `X-API-Key` header and work with most resources.

```bash
# Environment variable (recommended)
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
```

```hcl
# Or in provider configuration
provider "wayscloud" {
  api_key = var.wayscloud_api_key  # Use a variable, never hardcode!
}
```

### Personal Access Token (PAT)

PAT tokens are required for the `wayscloud_database` resource. The provider auto-detects the token type based on the prefix.

```bash
export WAYSCLOUD_API_KEY="wayscloud_pat_xxx..."  # PAT for database resource
```

| Resource | Auth Type | Required Scopes |
|----------|-----------|-----------------|
| `wayscloud_dns_zone` | API Key | `dns` |
| `wayscloud_dns_record` | API Key | `dns` |
| `wayscloud_redis_instance` | API Key | `redis` |
| `wayscloud_s3_bucket` | API Key | `storage` |
| `wayscloud_database` | **PAT** | `database:read`, `database:write` |
| `wayscloud_vps` | API Key | `vps` |
| `wayscloud_app` | API Key | `apps` |

### Base URL Override (Staging/Testing)

```bash
export WAYSCLOUD_ENDPOINT="https://api-staging.wayscloud.services"
```

```hcl
provider "wayscloud" {
  endpoint = "https://api-staging.wayscloud.services"
}
```

## Resources

| Resource | Description | Status |
|----------|-------------|--------|
| `wayscloud_dns_zone` | Manages DNS zones | Stable |
| `wayscloud_dns_record` | Manages DNS records (A, AAAA, CNAME, MX, TXT, etc.) | Stable |
| `wayscloud_redis_instance` | Redis as a Service | Stable |
| `wayscloud_s3_bucket` | S3-compatible object storage | Stable |
| `wayscloud_database` | Managed PostgreSQL/MariaDB | Stable |
| `wayscloud_vps` | Virtual Private Servers | Stable |
| `wayscloud_app` | Container app platform | Stable |

## Quick Start

```hcl
terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloud/wayscloud"
      version = "~> 0.1.0"
    }
  }
}

provider "wayscloud" {}

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

# Create a Redis cache
resource "wayscloud_redis_instance" "cache" {
  name   = "my-app-cache"
  region = "no"
  plan   = "redis-starter"
}
```

## Import Examples

All resources support importing existing infrastructure:

```bash
# DNS Zone (by domain name)
terraform import wayscloud_dns_zone.example example.com

# DNS Record (by zone/record_id)
terraform import wayscloud_dns_record.www example.com/550e8400-e29b-41d4-a716-446655440000

# Redis Instance (by UUID)
terraform import wayscloud_redis_instance.cache 550e8400-e29b-41d4-a716-446655440000

# S3 Bucket (by bucket name)
terraform import wayscloud_s3_bucket.uploads my-bucket-name

# Database (by type/name)
terraform import wayscloud_database.app postgresql/myapp-prod

# VPS (by UUID)
terraform import wayscloud_vps.web 550e8400-e29b-41d4-a716-446655440000

# App (by app ID)
terraform import wayscloud_app.api app_01ARZ3NDEKTSV4RRFFQ69G5FAV
```

## Known Limitations

### ForceNew Attributes

The following attributes require resource replacement (destroy + create):

| Resource | ForceNew Attributes |
|----------|---------------------|
| `wayscloud_dns_zone` | `name` |
| `wayscloud_dns_record` | `zone_name`, `name`, `type` |
| `wayscloud_redis_instance` | `name`, `region`, `plan` |
| `wayscloud_s3_bucket` | `bucket_name`, `tier` |
| `wayscloud_database` | `name`, `type`, `tier` |
| `wayscloud_vps` | `hostname`, `plan_code`, `region`, `os_template`, `ssh_keys` |
| `wayscloud_app` | `slug`, `region` |

### Async Resources

Some resources have async provisioning. The provider polls until ready:

| Resource | Typical Wait Time | Polling Interval |
|----------|-------------------|------------------|
| `wayscloud_redis_instance` | 2-5 minutes | 10 seconds |
| `wayscloud_vps` | 3-10 minutes | 15 seconds |

### Sensitive Attributes

These attributes are only available on initial creation and cannot be retrieved after import:

- `wayscloud_database.password`, `wayscloud_database.connection_string`
- `wayscloud_redis_instance.password`
- `wayscloud_s3_bucket.secret_key`

## Rate Limits & Retries

The provider automatically:
- Retries transient errors (429, 502, 503, 504) up to 3 times
- Uses exponential backoff between retries
- Times out requests after 30 seconds

## Development

### Building

```bash
go build -o terraform-provider-wayscloud
```

### Testing

```bash
# Unit tests
go test ./... -race

# Stress test
go test ./... -count=50

# Acceptance tests (requires credentials)
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
export TF_ACC=1
go test ./... -v -timeout 60m
```

### Parallel Apply Testing

```bash
# Test with high parallelism
terraform apply -parallelism=20 -auto-approve
terraform refresh
terraform plan  # Should show no changes
terraform destroy -parallelism=20 -auto-approve
```

## Security

See [SECURITY.md](SECURITY.md) for:
- How to report vulnerabilities
- Best practices for CI/CD
- Token handling guidelines

**Never commit API keys or PAT tokens to version control.**

## License

[Mozilla Public License 2.0](LICENSE)

## Support

- [WAYSCloud Documentation](https://docs.wayscloud.net)
- [GitHub Issues](https://github.com/wayscloud/terraform-provider-wayscloud/issues)
- [WAYSCloud Support](https://wayscloud.services/support)
