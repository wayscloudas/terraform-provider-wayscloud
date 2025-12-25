# Create a DNS zone for your domain
resource "wayscloud_dns_zone" "example" {
  name = "example.com"
}

# Output the nameservers to configure at your registrar
output "nameservers" {
  value       = wayscloud_dns_zone.example.nameservers
  description = "Configure these nameservers at your domain registrar"
}

output "zone_id" {
  value       = wayscloud_dns_zone.example.id
  description = "The zone UUID"
}
