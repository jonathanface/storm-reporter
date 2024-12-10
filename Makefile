# Variables
DOCKER_COMPOSE_FILE=docker-compose.yml
ETL_SERVICE=etl-service
KAFKA_BROKER=kafka:9092

# Default task: start all services
.PHONY: default
default: compose-up

# Build all Docker Compose services
.PHONY: compose-build
compose-build:
	docker compose -f $(DOCKER_COMPOSE_FILE) build

# Start Docker Compose services
.PHONY: compose-up
compose-up:
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d

# Stop Docker Compose services
.PHONY: compose-down
compose-down:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

# Restart Docker Compose services
.PHONY: compose-restart
compose-restart: compose-down compose-up

.PHONY: compose-replace
compose-replace: compose-down compose-build compose-up

.PHONY: force-publish
force-publish:
	docker exec -it producer-service npm run forcePublish

.PHONY: generate-storms
generate-storms:
	docker exec -it producer-service npm run generateStorms

.PHONY: drop-db
drop-db:
	docker volume rm weather_mongo_data

.PHONY: connect-db
connect-db: docker exec -it mongo mongosh

.PHONY: stop-container
stop-container:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please specify the container name using 'make stop-container name=<container_name>'"; \
		exit 1; \
	fi
	@docker stop $(name) && echo "Container '$(name)' stopped successfully." || echo "Failed to stop container '$(name)'."

.PHONY: start-container
start-container:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please specify the container name using 'make start-container name=<container_name>'"; \
		exit 1; \
	fi
	@docker start $(name) && echo "Container '$(name)' started successfully." || echo "Failed to start container '$(name)'."

.PHONY: container-status
container-status:
	@docker ps -a --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "Failed to retrieve container status."


# View logs of a specific service
.PHONY: logs
logs:
	@read -p "Enter service name (default: $(ETL_SERVICE)): " service && \
	docker compose -f $(DOCKER_COMPOSE_FILE) logs -f $${service:-$(ETL_SERVICE)}

# List Kafka topics
.PHONY: kafka-list
kafka-list:
	docker exec -it kafka kafka-topics --bootstrap-server $(KAFKA_BROKER) --list

# Create a Kafka topic
.PHONY: kafka-create
kafka-create:
	@read -p "Enter topic name: " topic && \
	docker exec -it kafka kafka-topics --bootstrap-server $(KAFKA_BROKER) --create --topic $$topic --partitions 1 --replication-factor 1

# Describe a Kafka topic
.PHONY: kafka-describe
kafka-describe:
	@read -p "Enter topic name: " topic && \
	docker exec -it kafka kafka-topics --bootstrap-server $(KAFKA_BROKER) --describe --topic $$topic

# Produce a message to a Kafka topic
.PHONY: kafka-produce
kafka-produce:
	@read -p "Enter topic name: " topic && \
	docker exec -it kafka kafka-console-producer --bootstrap-server $(KAFKA_BROKER) --topic $$topic

# Consume messages from a Kafka topic
.PHONY: kafka-consume
kafka-consume:
	@read -p "Enter topic name: " topic && \
	docker exec -it kafka kafka-console-consumer --bootstrap-server $(KAFKA_BROKER) --topic $$topic --from-beginning

# Help menu
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make start-container name=<container_name>   - Start a specific container"
	@echo "  make stop-container name=<container_name>   - Stop a specific container"
	@echo "  make container-status   - View service status"
	@echo "  make compose-build   - Build Docker Compose services"
	@echo "  make compose-up      - Start all services in Docker Compose"
	@echo "  make compose-down    - Stop all services in Docker Compose"
	@echo "  make compose-restart - Restart all services in Docker Compose"
	@echo "  make compose-replace - Stop all services, rebuild, and start all services in Docker Compose"
	@echo "  make logs            - View logs of a specific service"
	@echo "  make kafka-list      - List Kafka topics"
	@echo "  make kafka-create    - Create a Kafka topic"
	@echo "  make kafka-describe  - Describe a Kafka topic"
	@echo "  make kafka-produce   - Produce a message to a Kafka topic"
	@echo "  make kafka-consume   - Consume messages from a Kafka topic"
