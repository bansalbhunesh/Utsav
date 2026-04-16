.PHONY: docker-up docker-down api web

docker-up:
	docker compose -f infra/docker/compose.yml up -d

docker-down:
	docker compose -f infra/docker/compose.yml down

# From repo root on Git Bash / macOS / Linux:
api:
	cd services/api && \
	  MIGRATIONS_PATH=../../db/migrations \
	  DATABASE_URL=postgres://utsav:utsav@127.0.0.1:5432/utsav?sslmode=disable \
	  PORT=8080 \
	  HTTP_PORT=8080 \
	  go run ./cmd/api

web:
	cd frontend && npm install && npm run dev
