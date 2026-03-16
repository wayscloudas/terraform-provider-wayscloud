# Domain verification requires PAT auth (pat_token)
# Set WAYSCLOUD_PAT_TOKEN="wayscloud_pat_xxx..." or configure pat_token in provider

resource "wayscloud_domain_verification" "email" {
  domain              = "example.com"
  purpose             = "email"
  verification_method = "dns_txt"
}

# Create the verification DNS record (uses API key auth)
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
