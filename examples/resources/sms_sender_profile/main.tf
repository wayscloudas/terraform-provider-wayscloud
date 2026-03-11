resource "wayscloud_sms_sender_profile" "alerts" {
  name        = "Alert System"
  sender_id   = "WAYSCloud"
  allow_reply = false
  is_default  = true
}
