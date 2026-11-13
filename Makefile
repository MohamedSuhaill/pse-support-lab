COMPOSE ?= docker compose
CUSTOMER ?= desert-rose

.DEFAULT_GOAL := help
.PHONY: help up down reset load replay logs ps psql samples

help: ## Show available commands
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  make %-8s %s\n", $$1, $$2}'

up: ## Build and start the whole stack, wait until it is healthy
	$(COMPOSE) up -d --build --wait --wait-timeout 600
	@echo ""
	@echo "  Portal      http://localhost:3000"
	@echo "  Grafana     http://localhost:3001"
	@echo "  API         http://localhost:8080"
	@echo "  ASP (mock)  http://localhost:8090"
	@echo "  Prometheus  http://localhost:9090"

down: ## Stop the stack (keeps data)
	$(COMPOSE) down

reset: ## Delete all data (database, metrics, logs) and start again
	$(COMPOSE) down -v --remove-orphans
	$(MAKE) up

load: ## Replay this morning's customer traffic (takes a few minutes)
	$(COMPOSE) --profile tools run --rm --build --no-deps tools loadgen load

replay: ## Replay one customer upload: make replay CUSTOMER=desert-rose|oasis|al-noor|sahara|ghaf
	$(COMPOSE) --profile tools run --rm --build --no-deps tools loadgen replay $(CUSTOMER)

logs: ## Follow api and asp-mock logs
	$(COMPOSE) logs -f --tail=100 api asp-mock

ps: ## Show containers
	$(COMPOSE) ps

psql: ## Open a SQL shell on the portal database
	$(COMPOSE) exec postgres psql -U invoicely -d invoicely

samples: ## Regenerate the files in samples/
	$(COMPOSE) --profile tools run --rm --build --no-deps --user "$$(id -u):$$(id -g)" -v "$(CURDIR)/samples:/out" tools samplegen -out /out
