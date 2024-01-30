# Build and run the app image locally
run:
	$(MAKE) -C .local build
	docker volume create blinkfile-data
	docker run --name blinkfile -p 8000:8000 -e FEATURE_FLAG_LogAllAuthnCalls=1 -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=1234123412341234 -e DATA_DIR=/data -v blinkfile-data:/data --rm blinkfile-candidate
.PHONY: run

# Run full test suite, similar to what is run in CI pipeline
test:
	$(MAKE) -C .local all
.PHONY: test

# Install dependencies on host machine
install:
	go mod download
	cd app/web && npm i
.PHONY: install

# Run unit tests on the host machine
unit-test:
	go test -coverprofile coverage.out ./...
	go tool cover -html=coverage.out -o=coverage.html
.PHONY: unit-test

# Run linter
lint:
	$(MAKE) -C .local lint
.PHONY: lint

# Run acceptance tests
acceptance-test:
	$(MAKE) -C .local acceptance-test
.PHONY: acceptance-tests-runner

# Open acceptance test runner, requires a running app on http://localhost:8000
acceptance-test-runner:
	$(MAKE) -C test/ open
.PHONY: acceptance-test-runner

# Build github pages docs locally
build-docs:
	cd docs && bundle install
.PHONY: build-docs

# Run the github pages docs locally
run-docs:
	cd docs && bundle exec jekyll serve
.PHONY: run-docs

commitlint:
	$(MAKE) -C .local commitlint
.PHONY: commitlint