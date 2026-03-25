include .env

stop_containers:
	@echo "Stopping other docker containers"
	if [ $$(docker ps -q) ]; then \
  		echo "found and stopped containers"; \
  		docker stop $$(docker ps -q); \
	else \
		echo "no containers running..."; \
	fi

create_container:
	docker run --name ${DB_DOCKER_CONTAINER} -p 5432:5432 -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=${DB_PASSWORD} -d postgres:12-alpine

create_db:
	docker exec -it ${DB_DOCKER_CONTAINER} createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}

start_container:
	docker start ${DB_DOCKER_CONTAINER}

create_migrations:
	goose -s create name sql

migrate_up:
	goose up

migrate_down:
	goose down

build:
	if [ -f "${SERVER_BINARY}" ]; then \
		rm ${SERVER_BINARY}; \
		echo "Deleted ${SERVER_BINARY}"; \
	fi
	@echo "Building binary..."
	go build -o ${SERVER_BINARY} cmd/*.go

run: build
	./${SERVER_BINARY}

stop:
	@echo "stopping server..."
	@-pkill -SIGTERM -f "./${SERVER_BINARY}"
	@echo "server stopped..."