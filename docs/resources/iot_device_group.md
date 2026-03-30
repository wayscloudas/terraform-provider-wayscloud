---
page_title: "wayscloud_iot_device_group Resource - WAYSCloud"
subcategory: ""
description: |-
  Manages an IoT device group in WAYSCloud.
---

# wayscloud_iot_device_group (Resource)

Manages an IoT device group in WAYSCloud.

Device groups allow you to organize IoT devices into logical collections for bulk operations, monitoring rules, and access control.

## Example Usage

```hcl
resource "wayscloud_iot_device_group" "sensors" {
  name        = "Temperature Sensors"
  description = "All temperature sensors in building A"
}
```

## Argument Reference

- `name` - (Required) Human-readable name for the device group.
- `description` - (Optional) Description of the device group.

## Attribute Reference

- `id` - Unique identifier for the device group (UUID).
- `device_count` - Number of devices in the group.
- `created_at` - Timestamp when the group was created (ISO 8601).

## Import

IoT device groups can be imported using the group ID:

```bash
terraform import wayscloud_iot_device_group.sensors 550e8400-e29b-41d4-a716-446655440000
```
