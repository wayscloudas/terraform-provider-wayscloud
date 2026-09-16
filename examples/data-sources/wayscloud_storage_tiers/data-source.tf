data "wayscloud_storage_tiers" "all" {}

output "available_tiers" {
  value = data.wayscloud_storage_tiers.all.tiers
}
