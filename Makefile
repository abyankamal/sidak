.PHONY: help infra-up infra-down infra-logs dev-backend dev-web dev-mobile migrate-up migrate-down test-backend

help:
	@echo "=================================================================="
	@echo "SIDAK (Sistem Data Kewilayahan) - Development Commands"
	@echo "=================================================================="
	@echo "  make infra-up      - Start local Postgres 16 & Gotenberg containers"
	@echo "  make infra-down    - Stop local infra containers"
	@echo "  make infra-logs    - Follow logs from infra containers"
	@echo "  make migrate-up    - Run all database migrations (UP)"
	@echo "  make migrate-down  - Rollback database migrations (DOWN)"
	@echo "  make dev-backend   - Run Golang backend API server"
	@echo "  make dev-web       - Run Next.js web portal (Public & Admin)"
	@echo "  make dev-mobile    - Run Flutter mobile application"
	@echo "  make test-backend  - Run backend unit & integration tests with race detector"

infra-up:
	docker compose -f deploy/docker-compose.yml up -d

infra-down:
	docker compose -f deploy/docker-compose.yml down

infra-logs:
	docker compose -f deploy/docker-compose.yml logs -f

migrate-up:
	@echo "Applying database migrations..."
	@docker exec -i sidak_postgres psql -U postgres -d sidak_db < backend/migrations/000001_init_schema.up.sql
	@docker exec -i sidak_postgres psql -U postgres -d sidak_db < backend/migrations/000002_seed_initial_data.up.sql
	@echo "Migrations applied successfully."

migrate-down:
	@echo "Rolling back database migrations..."
	@docker exec -i sidak_postgres psql -U postgres -d sidak_db < backend/migrations/000002_seed_initial_data.down.sql
	@docker exec -i sidak_postgres psql -U postgres -d sidak_db < backend/migrations/000001_init_schema.down.sql
	@echo "Rollback completed."

test-backend:
	cd backend && go test -v -race ./...
