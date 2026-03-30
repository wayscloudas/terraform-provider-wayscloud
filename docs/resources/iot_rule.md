---
page_title: "wayscloud_iot_rule Resource - WAYSCloud"
subcategory: ""
description: |-
  Manages an IoT alerting rule in WAYSCloud.
---

# wayscloud_iot_rule (Resource)

Manages an IoT alerting rule in WAYSCloud.

IoT rules define conditions that trigger alerts based on device telemetry data. Rules can target an entire fleet, a device group, a device profile, or a single device.

## Example Usage

### Threshold Rule

```hcl
resource "wayscloud_iot_rule" "temp_alert" {
  name       = "High Temperature Alert"
  rule_type  = "threshold"
  scope_type = "fleet"
  severity   = "warning"

  config = jsonencode({
    field    = "temperature"
    operator = ">"
    value    = 80
    duration = 300
  })

  cooldown_seconds = 600
  auto_resolve     = true
}
```

### Missing Data Rule

```hcl
resource "wayscloud_iot_rule" "offline_check" {
  name       = "Device Offline Detection"
  rule_type  = "missing_data"
  scope_type = "group"
  severity   = "critical"

  config = jsonencode({
    timeout_seconds = 3600
    group_id        = wayscloud_iot_device_group.sensors.id
  })

  auto_resolve = true
}
```

## Argument Reference

- `name` - (Required) Human-readable name for the rule.
- `rule_type` - (Required) Type of rule: `missing_data`, `offline`, `threshold`, `message_rate`, `reconnect_rate`. Changing this forces a new resource.
- `scope_type` - (Required) Scope of the rule: `fleet`, `group`, `profile`, `device`.
- `severity` - (Required) Alert severity: `critical`, `warning`, `info`.
- `config` - (Required) Rule configuration as a JSON string. Structure depends on `rule_type`. Use `jsonencode()` to build.
- `is_enabled` - (Optional) Whether the rule is enabled. Default: `true`.
- `cooldown_seconds` - (Optional) Minimum seconds between repeated alerts for the same rule.
- `auto_resolve` - (Optional) Whether alerts auto-resolve when the condition clears. Default: `false`.

## Attribute Reference

- `id` - Unique identifier for the rule (UUID).
- `created_at` - Timestamp when the rule was created (ISO 8601).

## Import

IoT rules can be imported using the rule ID:

```bash
terraform import wayscloud_iot_rule.temp_alert 550e8400-e29b-41d4-a716-446655440000
```
