---
name: tt-deploy
description: Deploy 25types to production. Build, test, push, and verify.
user-invocable: true
---

Deploy 25types to the production server at 43.130.2.209.

## Steps

1. Run `go build ./...` from `backend/`. Stop and report if it fails.
2. Run `go test ./internal/...` from `backend/`. Ask before proceeding if tests fail.
3. Verify `.env` contains all required keys: CF_API_TOKEN, JWT_SECRET, RESEND_API_KEY, DODO_API_KEY, DODO_PRODUCT_MONTHLY, DODO_PRODUCT_YEARLY, DODO_WEBHOOK_KEY. List any missing keys and stop.
4. Run `scripts/deploy.sh`.
5. After deploy.sh finishes, run `scripts/smoke.sh https://25types.com`. Report passed/failed counts and any failures.
6. If all smoke tests pass, report "Deploy successful." If any fail, report the failures and suggest next steps.

Never skip a failed step without asking. The deploy.sh script handles: docker build, docker save, scp, docker load, docker compose up, health check loop.
