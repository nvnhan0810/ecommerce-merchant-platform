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
Merchants: `GET /api/v1/merchants` (admin JWT)

## Test

```bash
go test ./...
```

## Layout

```
cmd/server/
cmd/migrate/
internal/modules/<feature>/{domain,application,infrastructure,presentation}
internal/platform/{config,httpapi,postgres}
```

## Postgres + Goose migrations (Up / Down)

Set `DB_*` (or `DATABASE_URL`) in `.env`. On boot the API:

1. Connects with pgx pool
2. Runs Goose **Up** when `DB_AUTO_MIGRATE=true`
3. Seeds demo rows when `DB_SEED=true`

Migration files: `internal/platform/postgres/migrations/*.sql`

```sql
-- +goose Up
CREATE TABLE ...

-- +goose Down
DROP TABLE ...
```

### Refresh / rollback

Từ **host** (loopback + SSL; đừng đổi `api/.env` của pod):

```bash
make migrate-status
make migrate-up
make migrate-down          # rollback 1 bước
make migrate-refresh       # down all + up all — mất data
```

Sau `refresh`, restart API để seed lại:

```bash
make migrate-refresh
kubectl -n ecomerce-nvnhan0810-com rollout restart deployment/ecomerce-api
```

Tương đương:

```bash
cd api
DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/migrate refresh
```

> GIN là HTTP framework, không dùng để migrate. Stack hiện tại giữ Chi + Goose.
