# Domain verification requires PAT auth
provider "wayscloud" {
  alias   = "pat"
  api_key = var.wayscloud_pat_token # wayscloud_pat_xxx...
}

variable "wayscloud_pat_token" {
  type      = string
  sensitive = true
}

resource "wayscloud_domain_verification" "email" {
  provider            = wayscloud.pat
  domain              = "example.com"
  purpose             = "email"
  verification_method = "dns_txt"
}

# Create the verification DNS record (uses API key provider)
resource "wayscloud_dns_record" "verification" {
  zone_name = "example.com"
  name      = wayscloud_domain_verification.email.dns_record_name
  type      = "TXT"
  value     = wayscloud_domain_verification.email.dns_challenge
  ttl       = 300
}

output "verification_status" {
  value = wayscloud_domain_verification.email.status
}
