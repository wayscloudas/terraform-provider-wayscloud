terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloud/wayscloud"
      version = "0.4.0"
    }
  }
}

provider "wayscloud" {}

resource "wayscloud_domain_verification" "tf_test" {
  domain              = "tf-test-smoke-abc456.com"
  purpose             = "email"
  verification_method = "dns_txt"
}

output "domain_status" {
  value = wayscloud_domain_verification.tf_test.status
}

output "dns_challenge" {
  value = wayscloud_domain_verification.tf_test.dns_challenge
}
