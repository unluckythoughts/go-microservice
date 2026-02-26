.ONESHELL:
SHELL = /bin/bash

EXAMPLE_DOCKER_COMPOSE_FILE=example/docker-compose.yml
TEST_SETUP_DOCKERFILE=tests/setup/Dockerfile
TEST_SETUP_DOCKER_COMPOSE_FILE=tests/setup/docker-compose.yml
TEST_SETUP_IMAGE=test-service
TEST_SETUP_STACK_IMAGE=test-localstack:latest
DOCKER_COMPOSE ?= docker compose
CI_RUNNER_REPO = ci-runner-repo
DOCKERS_DIR = $(PWD)/dockers

DOCKERHUB_REGISTRY = unluckythoughts


check-db-ready:
	@echo -n "Waiting for db     ...  "
	while true; do \
		pg_id=$(shell ${DOCKER_COMPOSE} -f ${EXAMPLE_DOCKER_COMPOSE_FILE} ps -q example-db); \
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

init:
	go mod download
	go mod tidy

setup-example:
	@${DOCKER_COMPOSE} -f ${EXAMPLE_DOCKER_COMPOSE_FILE} kill 2>&1 1>/dev/null
	${DOCKER_COMPOSE} -f ${EXAMPLE_DOCKER_COMPOSE_FILE} rm -fsv 2>&1 1>/dev/null
	${DOCKER_COMPOSE} -f ${EXAMPLE_DOCKER_COMPOSE_FILE} up -d

test-setup-up:
	@${DOCKER_COMPOSE} -f ${TEST_SETUP_DOCKER_COMPOSE_FILE} up -d test-service test-db test-redis
	-@docker rm -f test-stack
	@docker run -d --name test-stack --hostname test-stack --network setup_default -p 4566:4566 -e DOCKER_HOST=unix:///var/run/docker.sock -e DEBUG=1 -e SERVICES=sqs -v /var/run/docker.sock:/var/run/docker.sock ${TEST_SETUP_STACK_IMAGE}

test-setup-down:
	-@docker rm -f test-stack
	@${DOCKER_COMPOSE} -f ${TEST_SETUP_DOCKER_COMPOSE_FILE} down -v --remove-orphans

test-setup-build:
	@docker build -f ${TEST_SETUP_DOCKERFILE} -t ${TEST_SETUP_IMAGE} .

test-setup-pull-deps:
	@docker pull postgres:latest
	@docker pull bitnami/redis:latest
	@docker pull localstack/localstack:latest || docker pull localstack/localstack:latest || docker pull localstack/localstack:latest
	@docker tag localstack/localstack:latest ${TEST_SETUP_STACK_IMAGE}

test-setup: test-setup-down test-setup-up

run-example:
	@echo "Starting service   ..."
	go run example/main.go

check-dependencies: check-db-ready check-stack-ready

start-example: setup check-dependencies run-example
