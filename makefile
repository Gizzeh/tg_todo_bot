start:
	docker compose build
	docker compose up

stop:
	docker compose down

start_db:
	docker compose up -d postgres

migrate_up:
	docker build . -t tg_todo_bot:latest
	docker run \
	--network=tg_todo_bot_network \
	--env-file ./.env \
	tg_todo_bot:latest /bin/sh -c "/tg_todo_bot migrate up"

migrate_down:
	docker build . -t tg_todo_bot:latest
	docker run \
	--network=tg_todo_bot_network \
	--env-file ./.env \
	tg_todo_bot:latest /bin/sh -c "/tg_todo_bot migrate down"

delete_db_data:
	sudo rm -rf ./docker/postgres/pgdata