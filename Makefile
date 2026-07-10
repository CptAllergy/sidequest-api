include .env

stop_containers:
	@echo "Stopping other docker containers"
	if [ $$(docker ps -q) ]; then \
  		echo "found and stopped containers"; \
  		docker stop $$(docker ps -q); \
	else \
		echo "no containers running..."; \
	fi

create_network:
	docker network create sidequest-net

create_db_container:
	docker run --name ${DB_DOCKER_CONTAINER} --network sidequest-net -p 5432:5432 -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=${DB_PASSWORD} -d postgres:18

create_db:
	docker exec -it ${DB_DOCKER_CONTAINER} createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}
	docker exec -it ${DB_DOCKER_CONTAINER} createdb --username=${DB_USER} --owner=${DB_USER} ${SUPERTOKENS_DB_NAME}

create_supertokens_container:
	docker run --name ${SUPERTOKENS_DOCKER_CONTAINER} --network sidequest-net -p 3567:3567 -e POSTGRESQL_CONNECTION_URI="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_DOCKER_CONTAINER}:${DB_PORT}/${SUPERTOKENS_DB_NAME}" -d supertokens/supertokens-postgresql

start_container:
	docker start ${DB_DOCKER_CONTAINER}
	docker start ${SUPERTOKENS_DOCKER_CONTAINER}

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

test:
	go test -v ./...