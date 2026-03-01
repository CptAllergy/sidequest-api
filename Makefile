include .env

stop_containers:
	@echo "Stopping other docker containers"
	if [ $$(docker ps -q) ]; then \
  		@echo "found and stopped containers"; \
  		docker stop $$(docker ps -q); \
	else \
		@echo "no containers running..."; \
	fi