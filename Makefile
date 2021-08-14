help: ## Show this help, targets are ordered by alphabet
	@echo 'usage: make [target] ...'
	@echo ''
	@echo 'targets:'
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sort | sed 's/:.*##/#/' | column -t -c 2 -s '#'

build: ## Re-build backend service if update code
	docker compose build

up: ## Run all services locally using Docker
	docker compose up -d

down: ## Clear all local services using Docker
	docker compose down

run: ## Run the app locally
	go run main.go

log: ## Log backend
	docker compose logs -f backend

dev-up: ## Start local environment for development
	cd devstack && docker compose up -d

dev-down: ## Shutdown the local environment
	cd devstack && docker compose down

.PHONY: test
test:
	go test ./test

test-docker:
	go test ./test -env=docker