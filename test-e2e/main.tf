terraform {
  required_providers {
    wayscloud = {
      source = "wayscloud/wayscloud"
    }
  }
}

provider "wayscloud" {
  # Credentials from env: WAYSCLOUD_API_KEY, WAYSCLOUD_PAT_TOKEN
  endpoint = "http://127.0.0.1:7001"
}

# ============ DATA SOURCES (read-only, safe) ============

data "wayscloud_regions" "all" {}

data "wayscloud_vps_plans" "all" {}

data "wayscloud_vps_os_templates" "all" {}

data "wayscloud_app_plans" "all" {}

data "wayscloud_dns_zones" "all" {}

data "wayscloud_database_types" "all" {}

data "wayscloud_redis_plans" "all" {}

data "wayscloud_storage_tiers" "all" {}

# ============ RESOURCES ============

resource "wayscloud_dns_zone" "test" {
  name = "tf-e2e-test-zone.com"
}

resource "wayscloud_dns_record" "test_a" {
  zone_name = wayscloud_dns_zone.test.name
  type      = "A"
  name      = "www"
  value     = "93.184.216.34"
  ttl       = 3600
}

# ============ OUTPUTS ============

output "regions_count" {
  value = length(data.wayscloud_regions.all.regions)
}

output "vps_plans_count" {
  value = length(data.wayscloud_vps_plans.all.plans)
}

output "os_templates_count" {
  value = length(data.wayscloud_vps_os_templates.all.templates)
}

output "app_plans_count" {
  value = length(data.wayscloud_app_plans.all.plans)
}

output "dns_zone_created" {
  value = wayscloud_dns_zone.test.name
}

output "dns_record_created" {
  value = "${wayscloud_dns_record.test_a.name}.${wayscloud_dns_record.test_a.zone_name}"
}
