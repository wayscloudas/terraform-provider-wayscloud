# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in this Terraform provider, please report it responsibly:

1. **DO NOT** open a public GitHub issue
2. Email security@wayscloud.no with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested fixes

We will respond within 48 hours and work with you to understand and address the issue.

## Security Best Practices

### Never Commit Credentials

```hcl
# BAD - Never do this
provider "wayscloud" {
  api_key = "wayscloud_api_abc123..."  # NEVER commit real keys!
}

# GOOD - Use environment variables
provider "wayscloud" {}
# Set WAYSCLOUD_API_KEY environment variable

# GOOD - Use Terraform variables
variable "wayscloud_api_key" {
  type      = string
  sensitive = true
}

provider "wayscloud" {
  api_key = var.wayscloud_api_key
}
```

### CI/CD Security

1. **Use secrets management**: Store API keys in your CI/CD platform's secret store (GitHub Secrets, GitLab CI Variables, etc.)

2. **Never log secrets**: The provider is designed to never log API keys or PAT tokens. If you see tokens in logs, report it as a security issue.

3. **Limit permissions**: Create API keys with only the permissions needed for your workflow.

4. **Rotate keys regularly**: Rotate API keys periodically, especially after team changes.

### State File Security

Terraform state files may contain sensitive information:

1. **Use remote state** with encryption (S3 + KMS, Terraform Cloud, etc.)
2. **Never commit state files** to version control
3. **Restrict access** to state storage

### Example `.gitignore`

```gitignore
# Terraform
*.tfstate
*.tfstate.*
.terraform/
.terraform.lock.hcl

# Secrets
*.tfvars
!example.tfvars

# Local overrides
override.tf
override.tf.json
*_override.tf
*_override.tf.json
```

## Security Features

### Token Masking

The provider automatically masks tokens in error messages. If an API error contains token patterns, the message is redacted.

### Retry with Backoff

The provider implements exponential backoff for transient errors (429, 502, 503, 504), reducing the risk of denial-of-service from aggressive retries.

### Sensitive Attributes

The following attributes are marked as `sensitive` in the schema and will not appear in logs or plan output:

- `wayscloud_database.password`
- `wayscloud_database.connection_string`
- `wayscloud_redis_instance.password`
- `wayscloud_s3_bucket.secret_key`

## Changelog

Security-related changes are documented in [CHANGELOG.md](CHANGELOG.md) with the `[SECURITY]` prefix.
