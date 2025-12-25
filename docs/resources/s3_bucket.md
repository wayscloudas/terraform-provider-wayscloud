---
page_title: "wayscloud_s3_bucket Resource - WAYSCloud"
description: |-
  Manages an S3-compatible storage bucket in WAYSCloud.
---

# wayscloud_s3_bucket (Resource)

Manages an S3-compatible storage bucket in WAYSCloud.

## Example Usage

```terraform
resource "wayscloud_s3_bucket" "uploads" {
  bucket_name = "my-app-uploads"
  tier        = "standard"
}

output "s3_endpoint" {
  value = wayscloud_s3_bucket.uploads.endpoint
}
```

## Schema

### Required

- `bucket_name` (String) - Bucket name. Must be globally unique. Changing this forces a new resource.

### Optional

- `tier` (String) - Storage tier: "standard" or "archive". Default: "standard". Changing this forces a new resource.

### Read-Only

- `id` (String) - The bucket UUID.
- `endpoint` (String) - S3 endpoint URL.
- `access_key` (String) - S3 access key.
- `secret_key` (String, Sensitive) - S3 secret key. Only available on creation.
- `created_at` (String) - Timestamp when created.

## Import

S3 buckets can be imported using the bucket name:

```bash
terraform import wayscloud_s3_bucket.uploads my-app-uploads
```

~> **Note:** The `secret_key` attribute will not be available after import.
