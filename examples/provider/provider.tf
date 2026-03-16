# Configure the WAYSCloud Provider
#
# Authentication:
# - api_key:   For DNS, VPS, Storage, Redis, IoT, SMS, Apps (env: WAYSCLOUD_API_KEY)
# - pat_token: For Database, Domain Verification (env: WAYSCLOUD_PAT_TOKEN)

terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloud/wayscloud"
      version = "~> 0.4"
    }
  }
}

# Option 1: Use environment variables (recommended)
# export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
# export WAYSCLOUD_PAT_TOKEN="wayscloud_pat_xxx..."
provider "wayscloud" {}

# Option 2: Use variables (good for CI/CD)
# variable "wayscloud_api_key" {
#   type      = string
#   sensitive = true
# }
#
# variable "wayscloud_pat_token" {
#   type      = string
#   sensitive = true
# }
#
# provider "wayscloud" {
#   api_key   = var.wayscloud_api_key
#   pat_token = var.wayscloud_pat_token
# }
