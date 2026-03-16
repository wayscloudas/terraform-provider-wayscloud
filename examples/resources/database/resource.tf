# Database resources require PAT auth (pat_token)
# Set WAYSCLOUD_PAT_TOKEN="wayscloud_pat_xxx..." with database:read + database:write scopes

# PostgreSQL database for web application
resource "wayscloud_database" "app" {
  name        = "myapp-production"
  type        = "postgresql"
  tier        = "standard"
  description = "Main application database"
}

# MariaDB database with encryption
resource "wayscloud_database" "secure" {
  name        = "secure-data"
  type        = "mariadb"
  tier        = "encrypted"
  description = "Database for sensitive data with encryption at rest"
}

# Output connection details
output "app_db_host" {
  value       = wayscloud_database.app.host
  description = "Database host endpoint"
}

output "app_db_port" {
  value       = wayscloud_database.app.port
  description = "Database port"
}

output "app_db_username" {
  value       = wayscloud_database.app.username
  description = "Database username"
}

output "app_db_password" {
  value       = wayscloud_database.app.password
  sensitive   = true
  description = "Database password (only available on creation)"
}

output "app_db_connection_string" {
  value       = wayscloud_database.app.connection_string
  sensitive   = true
  description = "Full connection string"
}
