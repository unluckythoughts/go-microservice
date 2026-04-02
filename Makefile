.ONESHELL:
SHELL = /bin/bash

DOCKER_COMPOSE_FILE=tests/setup/docker-compose.yml
DOCKERFILE=examples/microservice/Dockerfile
DOCKER_COMPOSE ?= docker compose

test-setup:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up --build -d

run: test-setup

stop:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

clean:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans	