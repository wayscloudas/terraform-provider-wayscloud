# Changelog

All notable changes to the WAYSCloud Terraform Provider are documented here.

## [0.4.0] - 2026-03-16

### Added
- **Dual auth with separate fields**: New `pat_token` provider field and `WAYSCLOUD_PAT_TOKEN` environment variable for PAT authentication. API key and PAT can now be configured simultaneously — no more provider aliases needed.
- **Client dual auth**: HTTP client sends both `X-API-Key` and `Authorization: Bearer` headers when both tokens are configured.

### Fixed
- **VPS ipv4_address crash**: `IPv4Address` object from asyncpg now correctly cast to string. Previously caused Pydantic V2 validation error after successful VM creation.
- **Database 401 errors**: Database and Domain Verification resources now work alongside API key resources without provider aliases.
- **VPS pricing field**: `monthly_price_nok` → `monthly_price` + `currency`. Previously always null due to JSON tag mismatch. Now returns price in customer's preferred currency (NOK, SEK, DKK, EUR).
- **VPS plans data source**: `code` → `plan_code`, `vcpu` → `cpu_cores`, `monthly_price_nok` → `monthly_price` + `currency`.
- **App plans data source**: `vcpu` → `cpu_cores`, `ram_mb` → `memory_mb`, `monthly_price_nok` → `monthly_price` + `currency`, added `disk_gb` and `hourly_rate`.
- **App min_instances default**: Changed from 1 to 0 to match API default.
- **S3/Database created_at**: Set timestamp on create to prevent "unknown value after apply" error.
- **Domain verification casing**: Normalize `DNS_TXT` → `dns_txt` to prevent state mismatch.
- **OS templates endpoint**: Fixed route conflict and async DB error.
- **Database node_id**: Fixed foreign key error — dynamic node selection based on db_type and tier.
- **PAT prefix**: Dashboard now generates `wayscloud_pat_` prefix (not `wayscloud_api_`) for PATs.

### Changed
- Documentation fully updated: all examples, README, and resource docs now reflect dual auth pattern and correct pricing fields.
- VPS examples updated to use current OS templates (`ubuntu-24.04`, `windows-server-2025`).
- Version constraint examples updated to `~> 0.4`.
- All 11 resources and 3 data sources verified end-to-end against live backend with real credentials.

### Breaking Changes
- `wayscloud_vps.monthly_price_nok` renamed to `wayscloud_vps.monthly_price` (new `currency` field added).
- `wayscloud_vps_plans` data source: `code` → `plan_code`, `vcpu` → `cpu_cores`, `monthly_price_nok` → `monthly_price` + `currency`.
- `wayscloud_app_plans` data source: `vcpu` → `cpu_cores`, `ram_mb` → `memory_mb`, `monthly_price_nok` → `monthly_price` + `currency`, new fields `disk_gb`, `hourly_rate`.

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
