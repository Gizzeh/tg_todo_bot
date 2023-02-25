start_db:
	docker compose --env-file ./.env -f ./docker/postgres/docker-compose.yml up -d

stop_db:
	docker compose -f ./docker/postgres/docker-compose.yml down

delete_db_data:
	sudo rm -rf ./docker/postgres/pgdata