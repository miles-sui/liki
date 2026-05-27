---
name: tt-gen-codes
description: Generate passport activation codes via admin API.
user-invocable: true
---

Generate redemption codes via `POST /api/admin/codes`.

## Steps

1. Ask the user for:
   - count (1-1000)
   - duration_d (days, 0=permanent)
   - max_uses (0=unlimited)
   - notes (optional)
2. Call `POST /api/admin/codes` with the payload. Use localhost:8080 or the running API server.
3. Display codes in a readable list. If count > 20, show first 10 and last 5.
4. Remind the user: admin endpoints have no auth in demo mode — these codes grant passport access to anyone who redeems them.
