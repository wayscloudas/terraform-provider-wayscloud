terraform {
  required_providers {
    wayscloud = { source = "wayscloud/wayscloud" }
  }
}

provider "wayscloud" {
  endpoint = "http://127.0.0.1:7001"
}

# DNS Zone
resource "wayscloud_dns_zone" "test" {
  name = "tf-hardening-test.com"
}

# DNS Record
resource "wayscloud_dns_record" "test_a" {
  zone_name = wayscloud_dns_zone.test.name
  type      = "A"
  name      = "www"
  value     = "93.184.216.34"
  ttl       = 3600
}

# S3 Bucket
resource "wayscloud_s3_bucket" "test" {
  bucket_name = "tf-hardening-test-bucket"
  tier        = "standard"
}

# S3 Bucket Key (sub-resource)
resource "wayscloud_s3_bucket_key" "test" {
  bucket_name = wayscloud_s3_bucket.test.bucket_name
  name        = "tf-test-key"
}

# Outputs for verification
output "dns_zone_id" {
  value = wayscloud_dns_zone.test.id
}
output "dns_record_id" {
  value = wayscloud_dns_record.test_a.id
}
output "s3_bucket_name" {
  value = wayscloud_s3_bucket.test.bucket_name
}
output "s3_key_id" {
  value = wayscloud_s3_bucket_key.test.id
}
output "s3_key_secret" {
  value     = wayscloud_s3_bucket_key.test.secret_key
  sensitive = true
}
