.ONESHELL:
PRIVATE_REPOS = github.com/investing-bot
SHELL = /bin/bash

EXAMPLE_DOCKER_COMPOSE_FILE=example/docker-compose.yml

check-db-ready:
	@for retry in $(shell seq 1 10); do \
		sleep 1; \
		pg_id=$(shell docker-compose -f ${EXAMPLE_DOCKER_COMPOSE_FILE} ps -q example-db); \
		if ! docker exec $${pg_id} pg_isready > /dev/null; then \
			continue; \
		fi; \
		echo "DB is ready!"; \
		break; \
	done; \

setup:
	docker-compose -f ${EXAMPLE_DOCKER_COMPOSE_FILE} rm -f
	docker-compose -f ${EXAMPLE_DOCKER_COMPOSE_FILE} up -d

setup-sqs:
	aws --region us-west-2 --endpoint-url=http://localhost:4566 sqs create-queue --queue-name example | jq

run-example:
	go run example/main.go

start-example: setup check-db-ready run-example