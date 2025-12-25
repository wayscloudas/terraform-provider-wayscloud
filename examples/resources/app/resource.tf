# Basic web application
resource "wayscloud_app" "web" {
  name   = "My Web App"
  region = "no"
  plan   = "app-basic"

  port              = 8080
  health_check_path = "/health"

  env_vars = {
    NODE_ENV      = "production"
    LOG_LEVEL     = "info"
    API_BASE_URL  = "https://api.example.com"
  }
}

# API backend with scale-to-zero
resource "wayscloud_app" "api" {
  name   = "API Backend"
  region = "no"
  plan   = "app-standard"

  port              = 3000
  health_check_path = "/healthz"

  min_instances         = 0
  max_instances         = 5
  scale_to_zero_enabled = true
  idle_timeout_minutes  = 10

  env_vars = {
    DATABASE_URL = wayscloud_database.app.connection_string
    REDIS_URL    = "redis://:${wayscloud_redis_instance.cache.password}@${wayscloud_redis_instance.cache.endpoint}:${wayscloud_redis_instance.cache.port}"
  }
}

# Worker process (always running)
resource "wayscloud_app" "worker" {
  name   = "Background Worker"
  region = "no"
  plan   = "app-basic"

  port              = 8080
  health_check_path = "/ready"

  min_instances = 1
  max_instances = 1

  env_vars = {
    WORKER_CONCURRENCY = "4"
    QUEUE_URL          = "amqp://rabbitmq.example.com"
  }
}

# Output app URLs
output "web_url" {
  value       = wayscloud_app.web.default_url
  description = "Web app URL"
}

output "api_url" {
  value       = wayscloud_app.api.default_url
  description = "API backend URL"
}

output "web_status" {
  value       = wayscloud_app.web.status
  description = "Web app status"
}

# Example: Add custom domain (via DNS record)
# resource "wayscloud_dns_record" "app_cname" {
#   zone_name = "example.com"
#   name      = "app"
#   type      = "CNAME"
#   value     = "myapp.apps.wayscloud.services"
#   ttl       = 300
# }
