# wayscloud_domain_verification

Manages a domain verification request in WAYSCloud.

~> This resource requires a **PAT token** (`wayscloud_pat_*`), not an API key. Use provider aliases if mixing auth types.

## Example Usage

```hcl
resource "wayscloud_domain_verification" "email" {
  domain              = "example.com"
  purpose             = "email"
  verification_method = "dns_txt"
}

# Create the DNS challenge record
resource "wayscloud_dns_record" "verification" {
  zone_name = "example.com"
  name      = wayscloud_domain_verification.email.dns_record_name
  type      = "TXT"
  value     = wayscloud_domain_verification.email.dns_challenge
  ttl       = 300
}
```

## Argument Reference

- `domain` - (Required, Forces new resource) Domain to verify.
- `purpose` - (Required, Forces new resource) Purpose: `email`, `dkim`, `dmarc`, `spf`, `general`.
- `verification_method` - (Optional, Forces new resource) Method: `dns_txt` (default) or `dns_cname`.
- `metadata` - (Optional) Key-value metadata map.

## Attribute Reference

- `id` - Unique identifier (UUID).
- `status` - Verification status: `pending`, `verified`, `failed`, `revoked`.
- `dns_challenge` - DNS challenge value to set.
- `dns_record_name` - DNS record name for the challenge.
- `verified_at` - Timestamp when verified (null until verified).

## Import

```bash
terraform import wayscloud_domain_verification.email VERIFICATION_UUID
```
