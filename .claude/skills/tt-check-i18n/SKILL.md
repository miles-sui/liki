---
name: tt-check-i18n
description: Check translation coverage in the built frontend.
user-invocable: true
---

Check translation coverage using `scripts/check-translations.js`.

## Steps

1. Check if the frontend build output (`dist/`) exists.
2. If not, tell the user to run `npm run build` from the frontend directory first.
3. If dist exists, run `node scripts/check-translations.js`.
4. Report:
   - Missing translation keys (referenced in HTML but not in translations.json)
   - Unused translation keys (in translations.json but never referenced)
   - Coverage percentage (used_keys / total_keys)
5. If no issues: "All translations present and used."
