# Container scripts
run: build
	docker volume create blinkfile-data
	docker run --name blinkfile -p 8000:8000 -e FEATURE_FLAG_LogAllAuthnCalls=1 -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=1234123412341234 -e DATA_DIR=/data -v blinkfile-data:/data --rm blinkfile
.PHONY: run

build:
	docker build --tag blinkfile .
.PHONY: build

CONTAINER_REGISTRY = docker.io/benjohns1
deploy: ci
	docker tag blinkfile ${CONTAINER_REGISTRY}/blinkfile
	docker push ${CONTAINER_REGISTRY}/blinkfile
.PHONY: deploy

ci: build
	$(MAKE) -C test/ ci
.PHONY: ci

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
