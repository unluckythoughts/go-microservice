.ONESHELL:
PRIVATE_REPOS = github.com/investing-bot
SHELL = /bin/bash

EXAMPLE_DOCKER_COMPOSE_FILE=example/docker-compose.yml
CI_RUNNER_REPO = ci-runner-repo
DOCKERS_DIR = $(PWD)/dockers

DOCKERHUB_REGISTRY = unluckythoughts


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

init:
	@rm -rf vendor
	go mod vendor -v

setup-example:
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


############################# CI Targets #############################
CI_RUNNER = docker run -i --rm \
	--net host \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v "${PWD}:/go/code" \
	-v "${HOME}/.aws:/root/.aws" \
	-v "${PWD}/github_key:/root/.ssh/id_rsa" \
	-v "${PWD}/known_hosts:/root/.ssh/known_hosts" \
	-e "GOPRIVATE=${GOPRIVATE}" \
	-e "GIT_USER=${GIT_USER}" \
	-e "GIT_TOKEN=${GIT_TOKEN}" \
	-w /go/code \
	${DOCKERHUB_REGISTRY}/$(CI_RUNNER_REPO):latest $(1)

.PHONY: ci
ci:
	@$(call CI_RUNNER,${step})

setup-git:
	git config --global url."https://$${GIT_USER}:$${GIT_TOKEN}@github.com".insteadOf "https://github.com"

setup: setup-git init

lint:
	golangci-lint run

docker-build-ci-runner:
	docker build -t "${CI_RUNNER_REPO}:latest" -f ${DOCKERS_DIR}/ci-runner.Dockerfile .

docker-build: docker-build-ci-runner

docker-push-ci-runner: docker-build-ci-runner
	docker tag "${CI_RUNNER_REPO}:latest" "${DOCKERHUB_REGISTRY}/${CI_RUNNER_REPO}:latest"
	docker push "${DOCKERHUB_REGISTRY}/${CI_RUNNER_REPO}:latest"

docker-push: docker-push-ci-runner
