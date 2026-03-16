# WAYSCloud Terraform / Provision API – Operational Hardening Checklist (MANDATORY)

We are now in hardening phase.
No change is considered complete until it is verified against the real backend and a real account.

This checklist must be executed before any internal push or provider release.

Status must be reported as:
- implemented
- verified
- fixed
- or failed

Never mark anything as complete without live verification.

---

## 1. Live authentication verification

Test with real credentials (my operational account or internal test account with same permissions).

Verify both:

- API key authentication
- PAT authentication

Check that headers are actually sent.

Must verify with curl:

API key:

curl -H "Authorization: Bearer <api_key>" ...

PAT:

curl -H "Authorization: Bearer <pat_token>" ...

Verify that:

- correct header is sent
- correct endpoint used
- no 401
- no 403
- no silent fallback

Specifically verify:

- databases (PAT)
- domain verification (PAT)
- VPS (API key)
- apps (API key)
- data sources

---

## 2. Terraform smoke tests (real account)

Use my account or internal test account.

For every changed resource or data source:

terraform init
terraform validate
terraform plan
terraform apply (if safe)
terraform plan again → must be NO DIFF

Must verify:

- provider does not crash
- response mapping works
- no panic
- no invalid type
- no unexpected diff
- no missing fields

Specifically check:

- VPS create → read → state mapping
- IPv4 fields are strings, not objects
- env_vars preserved
- secrets preserved
- database auth works
- import works

---

## 3. Response shape verification (backend contract)

For each changed endpoint:

curl real API response

Compare with provider structs.

Verify:

- field names match
- types match
- string vs object correct
- optional fields handled
- null handled
- empty handled

Especially verify:

- ipv4_address
- ipv6_address
- env_vars
- price fields
- status fields
- id fields

Provider must not crash on read.

---

## 4. Provider config verification

Test provider config with:

- only api_key
- only pat_token
- both api_key + pat_token

Verify:

- correct client used
- correct header sent
- no override bug
- no missing header

Databases must use PAT.
Data-plane must use API key.

---

## 5. Docs / README / examples sync

Docs must match actual behavior.

Check:

- auth method correct
- PAT mentioned where required
- registry namespace correct
- prices correct
- field names correct
- examples compile

Run:

terraform init
terraform validate

for every example.

If example fails → docs not updated.

---

## 6. Import / state / diff stability

For each resource:

- create
- import
- plan

Plan must be empty.

Verify:

- no RequiresReplace when not needed
- no default mismatch
- no missing fields
- no secrets lost

---

## 7. Cleanup / sweeper

All test resources must be removed.

Verify:

- no leftover VPS
- no leftover DB
- no leftover apps
- no leftover DNS

Test prefix must be used:

tf-test-
internal-test-
wc-test-

Never use real production names.

---

## 8. Report format (required)

Report must contain:

- what changed
- what tested
- what failed
- what fixed
- what verified live

If live verification not done:

Status = implemented but not verified

Never write "complete" without live verification.
