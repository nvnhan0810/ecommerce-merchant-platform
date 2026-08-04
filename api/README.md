# Ecomerce API (Go)

DDD + CQRS HTTP API for the ecomerce monorepo.

## Run

```bash
export PATH="$HOME/.local/go/bin:$PATH"
cp .env.example .env
go run ./cmd/server
```

Health: `GET /api/health`  
Products (public): `GET /api/v1/products`, `GET /api/v1/media/*`  
Products (admin JWT): `POST/PUT/DELETE /api/v1/products`, `GET /api/v1/products/{id}`, `POST/DELETE /api/v1/products/{id}/image`  
Image object key: `shops/{merchant_id}/products/{product_id}/{uuid}{ext}`  
Merchants: `GET /api/v1/merchants` (admin JWT)

## Test

```bash
go test ./...
```

## Layout

```
cmd/server/
cmd/migrate/
cmd/seed/
internal/modules/<feature>/{domain,application,infrastructure,presentation}
internal/platform/{config,httpapi,postgres,seed}
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

Sau `refresh`, chạy seeder (hoặc restart API với `DB_SEED=true`):

```bash
make migrate-refresh
make seed
# hoặc chỉ products (sau khi đã có merchants):
make seed-products
make seed-orders
```

Tương đương:

```bash
cd api
DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/migrate refresh
DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/seed
DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/seed products
DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/seed orders
```

### Demo seeder

`make seed` / `go run ./cmd/seed` đảm bảo (idempotent):

| Role | Email | Password |
|------|-------|----------|
| admin | `admin@ecomerce.local`, `ops@ecomerce.local` | `ADMIN_BOOTSTRAP_PASSWORD` (default `Admin@123456`) |
| merchant | `shop@`, `fashion@`, `tech@`, `home@ecomerce.local` | `Shop@123456` |
| user | `buyer@`, `an@`, `binh@`, `chi@ecomerce.local` | `Buyer@123456` |

Plus **20** demo products (5 / merchant) gắn `merchant_id` thật, và **7** demo orders (mỗi trạng thái một đơn: `new` / `paid` / `confirmed` / `shipping` / `succeeded` / `failed` / `cancelled`). Mỗi order có `code` unique 10 ký tự A–Z/0–9 để tracking, thuộc 1 user + 1 merchant; line items chỉ lấy sản phẩm của merchant đó.  
`make seed-products` / `make seed-orders` seed riêng từng phần. API boot với `DB_SEED=true` cũng gọi cùng seeder.

### Auth

| Portal | Login | Me |
|--------|-------|----|
| Admin | `POST /api/v1/auth/login` | `GET /api/v1/auth/me` (admin token) |
| Merchant | `POST /api/v1/auth/merchant/login` | `GET /api/v1/auth/merchant/me` (merchant token) |

Body login: `{"email":"...","password":"..."}` → `{access_token, token_type, user}`.

> GIN là HTTP framework, không dùng để migrate. Stack hiện tại giữ Chi + Goose.
