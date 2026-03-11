# wayscloud_iot_device

Manages an IoT device in WAYSCloud.

## Example Usage

```hcl
resource "wayscloud_iot_device" "sensor" {
  device_id   = "temp-sensor-01"
  name        = "Temperature Sensor #1"
  description = "Office temperature monitoring"
  device_type = "sensor"

  metadata = {
    location = "building-a"
    floor    = "2"
  }
}
```

## Argument Reference

- `device_id` - (Required, Forces new resource) User-defined unique device identifier.
- `name` - (Required) Human-readable device name.
- `description` - (Optional) Device description.
- `device_type` - (Optional) Device type classification (e.g., `sensor`, `gateway`, `actuator`).
- `metadata` - (Optional) Key-value metadata map.
- `is_active` - (Optional) Whether the device is active. Default: `true`.

## Attribute Reference

- `mqtt_username` - (Sensitive) MQTT username. Only available on initial creation.
- `mqtt_password` - (Sensitive) MQTT password. Only available on initial creation.

## Import

```bash
terraform import wayscloud_iot_device.sensor temp-sensor-01
```
