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
git clone https://github.com/wayscloudas/terraform-provider-wayscloud.git
cd terraform-provider-wayscloud
go build -o terraform-provider-wayscloud
```

## Authentication

The provider requires an API key for authentication. You can obtain one from the [WAYSCloud Dashboard](https://my.wayscloud.services) under **Account → API Keys**.

### Environment Variable (Recommended)

```bash
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
terraform plan
```

### Provider Configuration

```hcl
provider "wayscloud" {
  api_key = var.wayscloud_api_key
}
```

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

# Output nameservers
output "nameservers" {
  value = wayscloud_dns_zone.example.nameservers
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `wayscloud_dns_zone` | Manages DNS zones |
| `wayscloud_dns_record` | Manages DNS records (A, AAAA, CNAME, MX, TXT, etc.) |

## Data Sources

Coming soon.

## Roadmap

- [ ] `wayscloud_redis_instance` - Redis as a Service
- [ ] `wayscloud_s3_bucket` - S3-compatible object storage
- [ ] `wayscloud_database` - Managed PostgreSQL/MariaDB
- [ ] `wayscloud_vps` - Virtual Private Servers
- [ ] `wayscloud_app` - Container app platform

## Development

### Building

```bash
go build -o terraform-provider-wayscloud
```

### Testing

```bash
# Unit tests
go test ./...

# Acceptance tests (requires WAYSCLOUD_API_KEY)
TF_ACC=1 go test ./... -v
```

### Documentation

Documentation is generated from the schema using [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs):

```bash
go generate ./...
```

## License

[Mozilla Public License 2.0](LICENSE)

## Support

- [WAYSCloud Documentation](https://docs.wayscloud.net)
- [GitHub Issues](https://github.com/wayscloudas/terraform-provider-wayscloud/issues)
- [WAYSCloud Support](https://wayscloud.services/support)
