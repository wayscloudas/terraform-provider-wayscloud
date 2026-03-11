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

output "mqtt_username" {
  value = wayscloud_iot_device.sensor.mqtt_username
}

output "mqtt_password" {
  value     = wayscloud_iot_device.sensor.mqtt_password
  sensitive = true
}
