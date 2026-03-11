resource "wayscloud_sms_keyword" "help" {
  keyword            = "HELP"
  description        = "Help keyword for customer support"
  webhook_url        = "https://api.example.com/sms/inbound"
  auto_reply_enabled = true
  auto_reply_message = "Thank you for contacting us. We will respond shortly."
}
