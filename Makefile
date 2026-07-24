include .env

up:
	@echo "Starting containers in background..."
	docker compose -f docker/docker-compose.yml up -d

down:
	@echo "Stopping and removing containers..."
	docker compose -f docker/docker-compose.yml down

start_containers:
	@echo "Starting Docker Compose stack..."
	docker compose -f docker/docker-compose.yml start

stop_containers:
	@echo "Stopping Docker Compose stack..."
	docker compose -f docker/docker-compose.yml stop

create_migrations:
	goose -s create name sql

migrate_up:
	goose up

migrate_down:
	goose down

sqlc:
	sqlc generate

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