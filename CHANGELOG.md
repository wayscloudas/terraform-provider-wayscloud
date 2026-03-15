# Changelog

All notable changes to the WAYSCloud Terraform Provider are documented here.

## [0.3.1] - 2026-03-15

### Added
- **Data sources**: `wayscloud_regions`, `wayscloud_dns_zones`, `wayscloud_vps_plans`, `wayscloud_vps_os_templates`, `wayscloud_app_plans`
- **Diagnostics**: Structured error diagnostics with `apiDiagnostic` and `dataSourceDiagnostic` helpers
- **Wait helper**: Generic `WaitForState` with transient error tolerance (429/502/503/504) and 404 detection
- **VPS validation**: `ValidateConfig` warns on Windows template/plan mismatch
- **Database import**: Support `type/tier/name` format (backward-compatible with `type/name`)
- **Schema versioning**: `Version: 0` on all resource schemas for future state migrations
- **Acceptance test scaffolding**: Test helpers, contract tests, sweep cleanup
- **Makefile**: Build, test, and install targets

### Fixed
- **App resource**: Env vars perpetual diff — preserve `env_vars` across `mapResponseToState` calls
- **VPS resource**: Null pointer on optional computed fields — add `else` branches setting null for `DisplayName`, `OSTemplate`, `VCPU`, `RAMMB`, `DiskGB`, `MonthlyPriceNOK`, `ProvisionedAt`
- **Non-API errors**: Classify network errors (timeout, DNS, TLS, connection refused) with user-friendly messages

### Changed
- Go minimum version bumped to 1.25 (required by terraform-plugin-framework v1.19.0)
- terraform-plugin-framework upgraded to v1.19.0
- terraform-plugin-testing v1.15.0 added
- Version constraint examples updated from `~> 0.1` / `~> 0.2` to `~> 0.3`

## [0.3.0] - 2026-03-11

### Added
- **Dual authentication**: Auto-detect PAT (`wayscloud_pat_`) vs API key (`wayscloud_api_`) from token prefix
- **New resources**: `wayscloud_app`, `wayscloud_iot_device`, `wayscloud_sms_sender_profile`, `wayscloud_sms_keyword`, `wayscloud_domain_verification`
- **Client hardening**: Token sanitization in error messages, retry logic with exponential backoff

## [0.2.1] - 2025-12-25

### Added
- Initial public release
- Resources: `wayscloud_dns_zone`, `wayscloud_dns_record`, `wayscloud_redis_instance`, `wayscloud_s3_bucket`, `wayscloud_database`, `wayscloud_vps`
