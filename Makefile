.PHONY: up down up-integration down-integration

up:
	docker compose up --wait

down:
	docker compose down

up-integration:
	docker compose --profile integration up --wait

down-integration:
	docker compose --profile integration down

