export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# Buildx needs this
export DOCKER_CLI_EXPERIMENTAL=enabled
export COMPOSE_DOCKER_CLI_BUILD=1
#export DOCKER_DEFAULT_PLATFORM=linux/arm64

export UID=$(id -u)
export GID=$(id -g)

SHELL := /bin/bash


# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: _setup_buildx _dev_init  ## Builds the production binaries, and exits
	@docker compose -f docker/docker-compose.yml run  --entrypoint "bash -l /mnt/docker/build.sh" app-dev

.PHONY: dev_start
dev_start: _setup_buildx _dev_init ## Starts the development environment.
	# --no-cache
	@docker compose -f docker/docker-compose.yml build --build-arg UID=$$(id -u) --build-arg GID=$$(id -g) app-dev
	@docker compose -f docker/docker-compose.yml up  app-dev

.PHONY: dev_start_debug
dev_start_debug:  _setup_buildx  _dev_init ## Starts the development container with a shell.
	@echo "To start the app, execute /mnt/docker/rundev.sh"
	@docker container rm app-dev || true
	@docker compose -f docker/docker-compose.yml run -u root --entrypoint "/usr/bin/bash -l" app-dev


.PHONY: dev_shell
dev_shell:  ## Attaches a shell to the running development environment. (make dev_start needed for it)
	@docker compose  -f docker/docker-compose.yml exec app-dev bash -l

.PHONY: dev_root_shell
dev_root_shell:  ## Attaches a root shell to the running development environment. (make dev_start needed for it)
	@docker compose  -f docker/docker-compose.yml exec -u root app-dev bash -l

.PHONY: lint
lint: _lint_prep _lint_exec  ## Executes the linter in the dev env

.PHONY: _lint_prep
_lint_prep: _setup_buildx _dev_init
	@docker compose -f docker/docker-compose.yml build --build-arg UID=$$(id -u) --build-arg GID=$$(id -g) app-dev

.PHONY: _lint_exec
_lint_exec: _setup_buildx _dev_init
	@docker compose -f docker/docker-compose.yml run  --entrypoint "bash -l /mnt/docker/lint.sh" app-dev

.PHONY: _setup_buildx
_setup_buildx:
	@docker buildx create --name app-building-node --platform linux/amd64 --use --bootstrap || true

.PHONY: _dev_init
_dev_init:
	@mkdir /tmp/app-tmp > /dev/null 2>&1 || true
