# Container scripts
run: build
	docker volume create blinkfile-data
	docker run --name blinkfile -p 8000:8000 -e FEATURE_FLAG_LogAllAuthnCalls=1 -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=1234123412341234 -e DATA_DIR=/data -v blinkfile-data:/data --rm blinkfile
.PHONY: run

build:
	docker buildx build --tag blinkfile --cache-to type=gha --cache-from type-gha .
.PHONY: build

# Host machine scripts
test:
	go test -coverprofile coverage.out ./...
	go tool cover -html=coverage.out -o=coverage.html
.PHONY: test

test-acceptance:
	$(MAKE) -C test/ test
.PHONY: test-acceptance

test-acceptance-open:
	$(MAKE) -C test/ open
.PHONY: test-acceptance-open

install:
	go mod download
	cd app/web && npm i
.PHONY: install

# Run checks
lint: src-load
	docker run -t --rm -v blinkfile-src:/app -v blinkfile-lint-cache:/root/.cache -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v
.PHONY: lint

src-load:
	$(MAKE) -C .local/load load
.PHONY: src-load

src-rm:
	$(MAKE) -C .local/load rm
.PHONY: src-rm