# First, create the DNS zone
resource "wayscloud_dns_zone" "example" {
  name = "example.com"
}

# A record for www subdomain
resource "wayscloud_dns_record" "www" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}

# A record for root domain (@)
resource "wayscloud_dns_record" "root" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "" # Empty string = root domain
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}

# CNAME record for blog subdomain
resource "wayscloud_dns_record" "blog" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "blog"
  type      = "CNAME"
  value     = "example.com"
  ttl       = 3600
}

# MX record for email
resource "wayscloud_dns_record" "mx" {
  zone_name = wayscloud_dns_zone.example.name
  name      = ""
  type      = "MX"
  value     = "mail.example.com"
  ttl       = 3600
  priority  = 10
}

# TXT record for SPF
resource "wayscloud_dns_record" "spf" {
  zone_name = wayscloud_dns_zone.example.name
  name      = ""
  type      = "TXT"
  value     = "v=spf1 include:_spf.wayscloud.services ~all"
  ttl       = 3600
}

# Wildcard A record
resource "wayscloud_dns_record" "wildcard" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "*"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}
