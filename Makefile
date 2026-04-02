.ONESHELL:
SHELL = /bin/bash

DOCKER_COMPOSE_FILE=examples/microservice/docker-compose.yml
DOCKERFILE=examples/microservice/Dockerfile
DOCKER_COMPOSE ?= docker compose

test-setup: stop
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up --build -d

test:
	go test ./tests/... -v

run: test-setup

stop:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

clean:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans	