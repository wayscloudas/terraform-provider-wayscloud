# Configure the WAYSCloud Provider
#
# Authentication can be done via:
# 1. Environment variable: WAYSCLOUD_API_KEY (recommended)
# 2. Provider configuration (below)
# 3. Terraform variable

terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloud/wayscloud"
      version = "~> 0.3"
    }
  }
}

# Option 1: Use environment variable (recommended)
# export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
provider "wayscloud" {}

# Option 2: Use variable (good for CI/CD)
# variable "wayscloud_api_key" {
#   type      = string
#   sensitive = true
# }
#
# provider "wayscloud" {
#   api_key = var.wayscloud_api_key
# }
