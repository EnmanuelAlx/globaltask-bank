# ==========================================
# Development & Setup
# ==========================================

.PHONY: init up down restart logs ps test create-user

init: ## Step 1: Initialize Supabase, migrations and create user
	@./scripts/init-project.sh

up: ## Step 2: Build and start all services
	docker-compose up -d --build

down: ## Restart all services
	supabase stop --no-backup --workdir infra/
	docker-compose down -v

logs: ## Show logs
	docker-compose logs -f

test: ## Run API tests
	docker compose exec api go test ./... -v

create-user: ## Create a developer user via Supabase Admin API (Usage: make create-user EMAIL=user@example.com PASS=password123)
	@if [ -z "$(EMAIL)" ] || [ -z "$(PASS)" ]; then \
		echo "Usage: make create-user EMAIL=user@example.com PASS=password123"; \
		exit 1; \
	fi
	@cd backend && \
	GOTOOLCHAIN=local go run cmd/tools/create_user.go -email $(EMAIL) -password $(PASS) > user_output.txt
	@USER_ID=$$(grep "User ID:" backend/user_output.txt | cut -d ':' -f2 | tr -d ' ') && \
	if [ ! -z "$$USER_ID" ]; then \
		echo "Syncing User $$USER_ID to DB..."; \
		docker exec -t supabase_db_globaltask-bank psql -U postgres -c "INSERT INTO public.profiles (id, full_name, role) VALUES ('$$USER_ID', 'Dev User', 'ADMIN') ON CONFLICT DO NOTHING;"; \
	fi
	@rm backend/user_output.txt

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
