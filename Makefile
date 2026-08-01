.PHONY: api web merchant admin test

api:
	cd api && go run ./cmd/server

web:
	cd web && npm run dev

merchant:
	cd merchant && npm run dev

admin:
	cd admin && npm run dev

test:
	cd api && go test ./...
	cd web && npm test
	cd merchant && npm test
	cd admin && npm test
