data "wayscloud_database_types" "all" {}

output "available_types" {
  value = data.wayscloud_database_types.all.database_types
}
