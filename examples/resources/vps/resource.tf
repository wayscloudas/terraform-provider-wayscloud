# Ubuntu web server with tags and labels
resource "wayscloud_vps" "web" {
  hostname     = "web01.example.com"
  display_name = "Production Web Server"
  plan_code    = "NO-Start-Linux-2cpu-4096mb-30gb"
  region       = "NO"
  os_template  = "ubuntu-24.04"

  ssh_keys = [
    "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB... user@workstation"
  ]

  tags = ["web", "prod"]

  labels = {
    env  = "prod"
    role = "frontend"
    team = "platform"
  }
}

# Debian database server with tags and labels
resource "wayscloud_vps" "db" {
  hostname     = "db01.example.com"
  display_name = "Database Server"
  plan_code    = "NO-Premium-Linux-4cpu-8192mb-100gb"
  region       = "NO"
  os_template  = "debian-12"

  ssh_keys = [
    "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB... admin@company"
  ]

  tags = ["database", "prod"]

  labels = {
    env  = "prod"
    role = "database"
  }
}

# Windows Server (requires Windows-specific plan, no SSH keys)
resource "wayscloud_vps" "windows" {
  hostname     = "win01.example.com"
  display_name = "Windows Application Server"
  plan_code    = "NO-Medium-Windows-4cpu-4096mb-64gb"
  region       = "NO"
  os_template  = "windows-server-2025"
}

# Output connection info
output "web_server_ip" {
  value       = wayscloud_vps.web.ipv4_address
  description = "Web server public IP"
}

output "db_server_ip" {
  value       = wayscloud_vps.db.ipv4_address
  description = "Database server public IP"
}

output "web_server_status" {
  value       = wayscloud_vps.web.status
  description = "Web server status"
}

# Example: Configure DNS record for VPS
# resource "wayscloud_dns_record" "web_a" {
#   zone_name = "example.com"
#   name      = "web01"
#   type      = "A"
#   value     = wayscloud_vps.web.ipv4_address
#   ttl       = 300
# }
