# wayscloud_sms_sender_profile

Manages an SMS sender profile in WAYSCloud.

~> All attributes force recreation (no in-place updates supported).

## Example Usage

```hcl
resource "wayscloud_sms_sender_profile" "alerts" {
  name        = "Alert System"
  sender_id   = "WAYSCloud"
  allow_reply = false
  is_default  = true
}
```

## Argument Reference

- `name` - (Required, Forces new resource) Profile name for internal reference.
- `sender_id` - (Required, Forces new resource) Sender ID displayed to recipients.
- `allow_reply` - (Optional, Forces new resource) Whether recipients can reply. Default: `true`.
- `is_default` - (Optional, Forces new resource) Whether this is the default profile. Default: `false`.

## Attribute Reference

- `id` - Unique identifier (UUID).
- `approval_status` - Approval status: `pending`, `approved`, `rejected`.

## Import

```bash
terraform import wayscloud_sms_sender_profile.alerts PROFILE_UUID
```
