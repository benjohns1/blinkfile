#####################
# VALIDATION CHECKS #
#####################

# Run all validation checks locally
all: _source _lint _unit-test _build _acceptance-test
.PHONY: all

# Run linter
lint: _source
	$(MAKE) _lint
.PHONY: lint

# Run unit-tests
unit-test: _source
	$(MAKE) _unit-test
.PHONY: unit-test

# Build an app candidate
build: _source
	$(MAKE) _build
.PHONY: build

# Build and run acceptance tests
acceptance-test: _source _build
	$(MAKE) _acceptance-test
.PHONY: acceptance-test

# Clean up and remove any created volumes
clean:
	docker volume rm blinkfile-src
	docker volume rm blinkfile-cache
.PHONY: clean

##############################
# INDIVIDUAL STEPS           #
# Do not run these manually  #
##############################

# Load the application source, install dependencies, and copy source to a docker volume for use in subsequent steps
_source:
	docker build --tag blinkfile-src -f src.Dockerfile ..
	docker volume create blinkfile-src
	docker volume create blinkfile-cache
	docker run --name blinkfile-src -v blinkfile-src:/out_src -v blinkfile-cache:/out_cache --rm blinkfile-src
.PHONY: _source

# Lint using golangci-lint
_lint:
	docker run -t --rm -v blinkfile-src:/app -v blinkfile-cache:/root/.cache -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v
.PHONY: _lint

# Run go unit tests
_unit-test:
	docker build --tag blinkfile-unit-test -f unit-test.Dockerfile ..
	docker run --name blinkfile-unit-test -v blinkfile-src:/src -v blinkfile-cache:/cache --rm blinkfile-unit-test
.PHONY: _unit-test

# Build the go binary and image
_build:
	docker build --tag blinkfile-candidate -f build.Dockerfile --progress=plain ..
.PHONY: _build

# Run acceptance tests on a build image
_acceptance-test:
	docker build --tag blinkfile-acceptance-test-runner ../test/cypress
	docker compose -f ../test/docker-compose.yml up --exit-code-from acceptance-test-runner
	docker compose -f ../test/docker-compose.yml down
.PHONY: _acceptance-test
