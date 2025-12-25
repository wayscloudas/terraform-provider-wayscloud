# Standard S3 bucket for general storage
resource "wayscloud_s3_bucket" "uploads" {
  bucket_name = "my-app-uploads"
  tier        = "standard"
}

# Enterprise bucket for high-performance workloads
resource "wayscloud_s3_bucket" "data" {
  bucket_name = "production-data"
  tier        = "enterprise"
}

# Output connection details
output "uploads_endpoint" {
  value       = wayscloud_s3_bucket.uploads.endpoint
  description = "S3 endpoint URL"
}

output "uploads_access_key" {
  value       = wayscloud_s3_bucket.uploads.access_key
  description = "S3 access key ID"
}

output "uploads_secret_key" {
  value       = wayscloud_s3_bucket.uploads.secret_key
  sensitive   = true
  description = "S3 secret access key (only available on creation)"
}

output "uploads_region" {
  value       = wayscloud_s3_bucket.uploads.region
  description = "S3 storage region"
}

# Example: Configure AWS CLI for WAYSCloud S3
# aws configure set aws_access_key_id ${wayscloud_s3_bucket.uploads.access_key}
# aws configure set aws_secret_access_key ${wayscloud_s3_bucket.uploads.secret_key}
# aws configure set endpoint_url ${wayscloud_s3_bucket.uploads.endpoint}
# aws s3 ls s3://my-app-uploads/
