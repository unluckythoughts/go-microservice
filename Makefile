.ONESHELL:
PRIVATE_REPOS = github.com/investing-bot
SHELL = /bin/bash

EXAMPLE_DOCKER_COMPOSE_FILE=example/docker-compose.yml

check-db-ready:
	@echo -n "Waiting for db     ...  "
	while true; do \
		pg_id=$(shell docker-compose -f ${EXAMPLE_DOCKER_COMPOSE_FILE} ps -q example-db); \
		if docker exec $${pg_id} pg_isready > /dev/null; then \
			echo "done"; \
			break; \
		fi; \
		sleep 1; \
	done; \

check-stack-ready:
	@echo -n "Waiting for stack  ...  "
	while true; do \
		ready=$$(curl -s http://localhost:4566/health | jq '[.services[] == "running"]|all' 2>/dev/null); \
		if [[ $${ready} == "true" ]]; then \
			echo "done"; \
			break; \
		fi; \
		sleep 1; \
	done; \

setup:
	@docker-compose -f ${EXAMPLE_DOCKER_COMPOSE_FILE} kill 2>&1 1>/dev/null
	docker-compose -f ${EXAMPLE_DOCKER_COMPOSE_FILE} rm -fsv 2>&1 1>/dev/null
	docker-compose -f ${EXAMPLE_DOCKER_COMPOSE_FILE} up -d

setup-sqs:
	aws --region us-west-2 --endpoint-url=http://localhost:4566 sqs create-queue --queue-name example | jq

run-example:
	@echo "Starting service   ..."
	go run example/main.go

check-dependencies: check-db-ready check-stack-ready

start-example: setup check-dependencies run-example