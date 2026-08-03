.PHONY: api web merchant admin test migrate-up migrate-down migrate-refresh migrate-status migrate-version

api:
	cd api && go run ./cmd/server

web:
	cd web && npm run dev

merchant:
	cd merchant && npm run dev

admin:
	cd admin && npm run dev

# Host CLI: pg_hba blocks node IP without SSL — use loopback + require.
# k3s pods keep using api/.env (DB_HOST=192.168.100.28, DB_SSLMODE=disable).
migrate-up:
	cd api && DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/migrate up

migrate-down:
	cd api && DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/migrate down

migrate-refresh:
	cd api && DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/migrate refresh

migrate-status:
	cd api && DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/migrate status

migrate-version:
	cd api && DB_HOST=127.0.0.1 DB_SSLMODE=require go run ./cmd/migrate version

test:
	cd api && go test ./...
	cd web && npm test
	cd merchant && npm test
	cd admin && npm test
