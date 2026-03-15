# v0.3.1 Hardening Report

Internal report for maintainers. Not published to registry.

## Changes Summary

### Diagnostics Overhaul
- Replaced ad-hoc `resp.Diagnostics.AddError("Client Error", ...)` with structured `apiDiagnostic()` in all 5 data sources
- Non-API errors (network, DNS, TLS) now classified with user-friendly messages via `classifyNonAPIError()`
- Empty result warnings on regions, vps_plans, vps_os_templates, app_plans data sources

### State Management Fixes
- **App env_vars**: API does not echo env_vars on read; previous behavior caused perpetual diff. Now preserving plan/state env_vars across mapResponseToState.
- **VPS optional fields**: Missing `else` null branches for DisplayName, OSTemplate, VCPU, RAMMB, DiskGB, MonthlyPriceNOK, ProvisionedAt caused stale values to persist.

### Schema Versioning
- All 11 resources now have `Version: 0` in schema, enabling future state migrations without breaking existing state files.

### Wait/Poll Hardening
- Generic `WaitForState` function with configurable pending/target states
- Transient error tolerance: up to 2 consecutive 429/502/503/504 before failing
- 404 during polling treated as terminal failure (resource deleted)

### Import Enhancement
- Database import now supports `type/tier/name` (3-part) format
- Backward compatible with legacy `type/name` (2-part, assumes `standard` tier)

### VPS Validation
- `ValidateConfig` warns when Windows OS template used with non-Windows plan code

### Test Infrastructure
- Acceptance test helpers with provider factory
- Contract tests for provider instantiation
- Sweep function for test resource cleanup
- Individual acceptance tests for all 12 resource/data source types

## Files Added
- `internal/provider/diagnostics.go`
- `internal/provider/wait.go`
- `internal/provider/regions_data_source.go`
- `internal/provider/dns_zones_data_source.go`
- `internal/provider/vps_plans_data_source.go`
- `internal/provider/vps_os_templates_data_source.go`
- `internal/provider/app_plans_data_source.go`
- `internal/provider/acc_test_helpers.go`
- `internal/provider/contract_test.go`
- `internal/provider/sweep_test.go`
- `internal/provider/*_acc_test.go` (12 files)
- `CHANGELOG.md`
- `Makefile`
- `maintainers/hardening-report.md`

## Files Modified
- `go.mod`, `go.sum` (framework v1.19.0, testing v1.15.0, go 1.25)
- `internal/provider/provider.go` (register data sources, version ref)
- `internal/provider/app_resource.go` (env_vars fix, Version: 0)
- `internal/provider/vps_resource.go` (null branches, ValidateConfig, Version: 0)
- `internal/provider/database_resource.go` (3-part import, Version: 0)
- All other resource files (Version: 0)
- `README.md`, `docs/index.md`, `examples/provider/provider.tf` (version refs)
