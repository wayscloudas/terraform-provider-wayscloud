data "wayscloud_redis_plans" "all" {}

output "available_plans" {
  value = data.wayscloud_redis_plans.all.plans
}
