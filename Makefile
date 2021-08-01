build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

log:
	docker compose logs -f backend