---
page_title: "wayscloud_s3_bucket_key Resource - WAYSCloud"
subcategory: ""
description: |-
  Manages an S3 bucket access key in WAYSCloud.
---

# wayscloud_s3_bucket_key (Resource)

Manages an S3 bucket access key in WAYSCloud.

Access keys provide programmatic access to S3 buckets. Each key has an access key ID and a secret key. The secret key is only available at creation time and cannot be retrieved later.

## Example Usage

```hcl
resource "wayscloud_s3_bucket" "data" {
  name = "my-data-bucket"
}

resource "wayscloud_s3_bucket_key" "app" {
  bucket_name = wayscloud_s3_bucket.data.name
  name        = "app-key"
}

output "access_key" {
  value = wayscloud_s3_bucket_key.app.access_key
}

output "secret_key" {
  value     = wayscloud_s3_bucket_key.app.secret_key
  sensitive = true
}
```

## Argument Reference

- `bucket_name` - (Required) The S3 bucket name this key belongs to. Changing this forces a new resource.
- `name` - (Required) Human-readable name for the access key. Changing this forces a new resource.

## Attribute Reference

- `id` - Unique identifier for the bucket key (UUID).
- `access_key` - The access key ID for S3 authentication.
- `secret_key` - The secret key for S3 authentication. Only available on initial creation.
- `created_at` - Timestamp when the key was created (ISO 8601).

## Import

S3 bucket keys can be imported using the format `bucket_name/key_id`:

```bash
terraform import wayscloud_s3_bucket_key.app my-data-bucket/550e8400-e29b-41d4-a716-446655440000
```

~> **Note:** The secret key cannot be retrieved after import. The value in state will be empty.
