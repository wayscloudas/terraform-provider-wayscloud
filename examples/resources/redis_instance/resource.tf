# Basic Redis instance for caching
resource "wayscloud_redis_instance" "cache" {
  name   = "my-app-cache"
  region = "no"
  plan   = "redis-starter"
}

# Redis with persistence for session storage
resource "wayscloud_redis_instance" "sessions" {
  name        = "session-store"
  region      = "no"
  plan        = "redis-standard"
  persistence = true
}

# Output connection details
output "cache_endpoint" {
  value       = wayscloud_redis_instance.cache.endpoint
  description = "Redis cache endpoint"
}

output "cache_port" {
  value       = wayscloud_redis_instance.cache.port
  description = "Redis cache port"
}

output "cache_password" {
  value       = wayscloud_redis_instance.cache.password
  sensitive   = true
  description = "Redis cache password (sensitive)"
}

# Example connection string for application
output "cache_connection_url" {
  value       = "redis://:${wayscloud_redis_instance.cache.password}@${wayscloud_redis_instance.cache.endpoint}:${wayscloud_redis_instance.cache.port}"
  sensitive   = true
  description = "Full Redis connection URL"
}
