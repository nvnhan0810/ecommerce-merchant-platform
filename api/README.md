# Ecomerce API (Go)

DDD + CQRS HTTP API for the ecomerce monorepo.

## Run

```bash
export PATH="$HOME/.local/go/bin:$PATH"
cp .env.example .env
go run ./cmd/server
```

Health: `GET /api/health`  
Products: `GET /api/v1/products`  
Merchants: `GET /api/v1/merchants`

## Test

```bash
go test ./...
```

## Layout

```
cmd/server/
internal/modules/<feature>/{domain,application,infrastructure,presentation}
internal/platform/{config,httpapi}
```

## Postgres

Set `DB_*` (or `DATABASE_URL`) in `.env`. On boot the API:

1. Connects with pgx pool
2. Runs embedded SQL migrations when `DB_AUTO_MIGRATE=true`
3. Seeds demo rows when `DB_SEED=true` and tables are empty
