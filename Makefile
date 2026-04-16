.PHONY: docker-up docker-down api web

docker-up:
	docker compose -f infra/docker/compose.yml up -d

docker-down:
	docker compose -f infra/docker/compose.yml down

# From repo root on Git Bash / macOS / Linux:
api:
	cd services/api && \
	  MIGRATIONS_PATH=../../db/migrations \
	  DATABASE_URL="$${DATABASE_URL:?DATABASE_URL is required}" \
	  PORT=8080 \
	  HTTP_PORT=8080 \
	  go run ./cmd/api

web:
	cd frontend && npm install && npm run dev
