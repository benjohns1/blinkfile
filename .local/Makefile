# Run all CI checks locally
ci: step-source step-lint
.PHONY: ci

# Load the application source, install dependencies, and copy source to a docker volume for use in subsequent steps
step-source:
	docker volume create blinkfile-src
	docker build --tag blinkfile-src-load -f load.Dockerfile ..
	docker run --name blinkfile-src-load -v blinkfile-src:/src --rm blinkfile-src-load
.PHONY: step-source

clean:
	docker volume rm blinkfile-src
	docker volume rm blinkfile-lint-cache
.PHONY: clean

step-lint:
	docker run -t --rm -v blinkfile-src:/app -v blinkfile-lint-cache:/root/.cache -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v
.PHONY: step-lint

step-unit-test:

.PHONY: step-unit-test