.PHONY: docker-up docker-down api web dev-api dev-web

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

# One-file local env flow: keep all vars in repo-root .env.local
dev-api:
	test -f .env.local || (echo ".env.local not found at repo root" && exit 1)
	set -a; . ./.env.local; set +a; cd services/api; go run ./cmd/api

dev-web:
	test -f .env.local || (echo ".env.local not found at repo root" && exit 1)
	set -a; . ./.env.local; set +a; cd frontend; npm install && npm run dev
