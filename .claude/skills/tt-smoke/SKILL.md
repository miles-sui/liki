---
name: tt-smoke
description: Run E2E smoke tests against local or production.
user-invocable: true
argument-hint: [base_url]
---

Run the full E2E smoke test suite via `scripts/smoke.sh`.

## Steps

1. If the user provides a base URL, use it. Otherwise default to `http://localhost`.
2. Run `scripts/smoke.sh [base_url]`.
3. Parse output and report:
   - Passed/failed counts per phase (phases 0-9)
   - Failed test details (endpoint, expected vs actual, error code)
   - Overall pass/fail summary
4. Don't dump raw curl output unless a test failed and the user asks for details.
