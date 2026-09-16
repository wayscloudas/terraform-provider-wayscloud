data "wayscloud_regions" "all" {}

# Codes of the regions that are currently active
output "active_regions" {
  value = [for r in data.wayscloud_regions.all.regions : r.code if r.available]
}

# Codes of the regions where App Platform can create apps
output "app_regions" {
  value = [for r in data.wayscloud_regions.all.regions : r.code if contains(r.available_services, "apps")]
}
