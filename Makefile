.PHONY: up down dev migrate migrate-ai seed generate dev-go dev-ai dev-web

up:
	docker-compose -f infra/docker/docker-compose.yml up -d

down:
	docker-compose -f infra/docker/docker-compose.yml down

dev:
	$(MAKE) -j3 dev-go dev-ai dev-web

migrate:
	@echo "Running Go migrations..."
	# TODO: execute golang-migrate script here
	./scripts/db/migrate.sh

migrate-ai:
	@echo "Running FastAPI migrations..."
	# TODO: execute alembic script here
	./scripts/db/migrate-ai.sh

seed:
	@echo "Seeding databases..."
	./scripts/dev/seed.sh

generate:
	@echo "Generating API contracts..."
	./scripts/generate.sh

dev-go:
	cd services/api-go && go run ./cmd/api

dev-ai:
	cd services/ai-fastapi && uvicorn app.main:app --reload

dev-web:
	cd apps/web && npm run dev

test:
	./scripts/test.sh

lint:
	./scripts/lint.sh
