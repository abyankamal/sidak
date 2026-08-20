.PHONY: help dev-backend dev-web dev-mobile migrate-up migrate-down test-backend

help:
	@echo "SIDAK Monorepo Development Commands"
	@echo "  make dev-backend   - Run Golang backend server"
	@echo "  make dev-web       - Run Next.js web (public + admin)"
	@echo "  make dev-mobile    - Run Flutter mobile app"
	@echo "  make test-backend  - Run backend unit tests"
	@echo "  make migrate-up    - Run database migrations UP"
	@echo "  make migrate-down  - Run database migrations DOWN"

test-backend:
	cd backend && go test -v -race ./...
