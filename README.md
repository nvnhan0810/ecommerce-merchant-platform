# Ecomerce monorepo

Một repo git — 4 source tách biệt FE/BE cho sàn thương mại điện tử.

| Source | Stack | Domain | Thư mục |
|--------|-------|--------|---------|
| User storefront | React + TypeScript (Vite) | `ecomerce.nvnhan0810.com` | `web/` |
| Merchant portal | Vue 3 + TypeScript (Vite) | `ecomerce-merchant.nvnhan0810.com` | `merchant/` |
| Admin console | Vue 3 + TypeScript (Vite) | `ecomerce-admin.nvnhan0810.com` | `admin/` |
| API | Go (Chi) | `ecomerce-api.nvnhan0810.com` | `api/` |

## Architecture

Tuân thủ Cursor global rules:

- **Frontend**: Clean Architecture — `Presentation → Application → Domain`, `Infrastructure → Domain`
- **Backend**: DDD + CQRS — `modules/<feature>/{domain,application,infrastructure,presentation}`
- **Testing**: unit / feature / component bắt buộc theo layer

## Quick start (local)

```bash
# API
export PATH="$HOME/.local/go/bin:$PATH"
cd api && go run ./cmd/server

# Storefront (port 5173)
cd web && cp .env.example .env && npm run dev

# Merchant (port 5174)
cd merchant && cp .env.example .env && npm run dev

# Admin (port 5175)
cd admin && cp .env.example .env && npm run dev
```

Dev proxy: mỗi FE Vite proxy `/api` → `http://127.0.0.1:8080`.  
Prod: set `VITE_API_BASE_URL=https://ecomerce-api.nvnhan0810.com`.

## API bootstrap

- `GET /api/health`
- `GET /api/v1/products`
- `POST /api/v1/products`
- `GET /api/v1/merchants`
- `GET /api/v1/users`

Store hiện tại là in-memory (seed demo). Có thể thay bằng Postgres trong `infrastructure/` mà không đụng domain.

## Tests

```bash
cd api && go test ./...
cd web && npm test
cd merchant && npm test
cd admin && npm test
```

## Docker

Mỗi source có `Dockerfile` riêng (FE → nginx static, API → Go binary).

## Postgres

API dùng PostgreSQL. Copy và điền credential:

```bash
cp api/.env.example api/.env
# DB_HOST / DB_PORT / DB_DATABASE / DB_USERNAME / DB_PASSWORD
```

Schema migrate tự chạy khi `DB_AUTO_MIGRATE=true`.

## Deploy k3s

```bash
cd ~/www/nvnhan0810-k3s/apps/ecomerce.nvnhan0810.com
./install.sh
```

Chi tiết: `nvnhan0810-k3s/apps/ecomerce.nvnhan0810.com/README.md`
