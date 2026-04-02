.ONESHELL:
SHELL := $(or $(shell command -v bash),$(shell command -v sh))

DOCKER_COMPOSE_FILE=examples/microservice/docker-compose.yml
DOCKERFILE=examples/microservice/Dockerfile
DOCKER_COMPOSE ?= docker compose

test-setup: stop
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up --build -d

test-ci: test-setup
	docker build -f tests/Dockerfile -t go-microservice-test-ci .
	docker run --rm \
	  --network go-microservice-test \
	  --env SERVICE_ENDPOINT_URL=http://service:8080/api/v1/ \
	  --env SERVICE_DB_HOST=db \
	  --name test-ci go-microservice-test-ci test

test:
	go test ./tests/... -v

test-auth:
	@if [ -n "$(test)" ]; then \
		go test ./tests/auth/... -v -run $(test); \
	else \
		go test ./tests/auth/... -v; \
	fi

run: test-setup

stop:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

clean:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans	