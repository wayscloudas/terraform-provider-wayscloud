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

# Example: Use with application deployment
# resource "wayscloud_app" "myapp" {
#   name = "my-application"
#
#   environment = {
#     DATABASE_URL = wayscloud_database.app.connection_string
#   }
# }
