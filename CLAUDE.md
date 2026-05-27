# 25types (Fivefold Types)

Personality assessment platform. Static HTML frontend + Go JSON API backend. SQLite. DDD.

## Build & Run

```
# Build all Go code
go build ./...

# Dev server — Docker (API :8080, Caddy :80)
scripts/dev.sh

# Dev server — local, no Docker (API :8081, Caddy :8080)
scripts/dev-local.sh

# Pre-commit check (no server needed)
make check

# Full local test suite (server must be running)
scripts/test-all.sh                   # unit + integ + frontend + smoke
scripts/test-all.sh --e2e             # + Playwright E2E

# Deploy to production
scripts/deploy.sh                     # build → scp → docker compose up
scripts/smoke.sh https://25types.com  # post-deploy journey smoke
```

## Architecture

```
cmd/
  app-server/    API server entry point
  mingli-server/ Standalone mingli API server
internal/
  ganzhi/        Stem-Branch calendar engine
  tianwen/       Chinese astronomy/calendar
  25types/       Personality computation engine
  httputil/      HTTP envelope helpers
  mingli/        Mingli domain
    bazi/        BaZi chart, match, fortune engines
    http/        Mingli API handlers (stateless)
    huangli/     Almanac, jieqi, mansions
    fengshui/    Fengshui computation
    qiming/      Naming engine
  app/           Main application (DDD)
    domain/      Entities & value objects
    application/ Use cases & ports
    sqlite/      Repository implementations
    http/        Handlers + middleware, server.go
    infra/       External clients (dodo/, resend/, tencent/)
    db/          SQLite open + migrations (embed)
data/            SQLite database files (dev only — production uses Docker volume)
deploy/          Dockerfiles and compose files
  app/           Main app server + Caddy
  caddy/         Caddy configs and Dockerfile
  mingli/        Standalone mingli server + Caddy
web/             Eleventy static frontend
```

Payments are provider-agnostic: `application/commerce/ports.go` defines `CommerceRepository`, implementations live in `sqlite/commerce_repo.go`. Currently using Dodo Payments (`infra/dodo/`).

## Key conventions

- Envelope: `{"data":{...}}` (single), `{"data":{"items":[...],"total":N}}` (list), `{"error":{"code":"...","message":"..."}}` (error)
- JWT: HS256, 30 days, env `JWT_SECRET`. Token invalidation via `token_version`.
- Sentinels: `domain.ErrXxx` — use `errors.Is()`, never string comparison.
- SQLite WAL mode, single connection (`MaxOpenConns=1`), forward-only migrations.
- API routes: no `/v1/` prefix. Caddy handles rate limits, TLS, static files.

## Docs entry point

`docs/INDEX.md` — architecture, ADRs, shared kernel, conventions. Always read first.
`docs/API.md` — HTTP contracts (the authoritative source).
`docs/OPERATIONS.md` — deploy, backups, email, payments, security.

## Workflow

Doc-first: write the doc, then write the code. Doc is spec, code is implementation.

Plan before every non-trivial implementation. Doc says what, plan says how. Enter plan mode, explore relevant existing code, produce a step-by-step plan, get approval, then execute. Never jump straight from doc to code — the plan is the bridge.

### Implementation discipline

Before starting any implementation, read the relevant `domain/*.md` into context. Never implement from memory.

**Start every module the same way:**
1. Read the domain doc — types, behavior rules, hard gates, function signatures are all there
2. Translate types — doc types → Go structs in `internal/app/domain/` or `internal/ganzhi/`/`internal/tianwen/`
3. Write function stubs — signatures must match the doc exactly (same params, same returns, no extras)
4. Implement the pipeline — follow the doc's step order, don't reorder or skip steps
5. Hard gates as explicit checks — doc says "六冲淘汰" → `if zhiRelation == "六冲" { continue }`

**Guardrails:**
- Function signatures are contracts. Don't add params, don't change return types. The doc is authoritative.
- No scoring unless the doc defines it. naming.md and dates.md explicitly reject manual scoring — don't invent weights.
- Data sources are specified in the doc. Don't pull in new dependencies without updating the doc first.
- If you find a gap in the doc, update the doc first, then change the code.

**Self-check after each module:**
1. Run `go build ./...` to confirm types compile
2. Re-read the doc's behavior rules section — verify each rule is enforced in code
3. Check hard gates are not skipped (六冲, 建除=破, etc.)
4. Verify no subjective scoring snuck in

### Doc → code mapping

| Doc section | Maps to |
|-------------|---------|
| §2 领域类型 | `internal/app/domain/` types, JSON tags must match |
| §3 行为规则 | `internal/ganzhi/`, `internal/tianwen/` pure functions |
| §4 数据 / Engine 签名 | `internal/ganzhi/`, `internal/mingli/` function stubs, `configs/` data files |
| §5 API 契约 → `docs/API.md` | `internal/app/http/`, `internal/mingli/http/` handlers, request/response types |

## Don't

- Don't change the `docs/` directory structure.
- Don't embed design knowledge in code comments — put it in docs/.
- Don't skip webhook signature verification.

## When removing a feature

- Grep `scripts/` for references (smoke tests, contract tests, E2E) and remove them in the same commit.
