# wayscloud_sms_keyword

Manages an SMS keyword in WAYSCloud for inbound message handling.

## Example Usage

```hcl
resource "wayscloud_sms_keyword" "help" {
  keyword            = "HELP"
  description        = "Help keyword for customer support"
  webhook_url        = "https://api.example.com/sms/inbound"
  auto_reply_enabled = true
  auto_reply_message = "Thank you for contacting us."
}
```

## Argument Reference

- `keyword` - (Required, Forces new resource) The keyword to match in inbound SMS messages.
- `description` - (Optional) Description of the keyword's purpose.
- `webhook_url` - (Optional) URL to receive webhook notifications.
- `auto_reply_enabled` - (Optional) Enable automatic reply. Default: `false`.
- `auto_reply_message` - (Optional) Auto-reply message text.
- `is_active` - (Optional) Whether the keyword is active. Default: `true`.

## Attribute Reference

- `id` - Unique identifier (UUID).

## Import

```bash
terraform import wayscloud_sms_keyword.help KEYWORD_UUID
```
