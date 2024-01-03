run: build
	docker volume create blinkfile-data
	docker run --name blinkfile -p 8000:8000 -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=1234123412341234 -e DATA_DIR=/data -v blinkfile-data:/data --rm blinkfile

build:
	docker build --tag blinkfile .

deploy: build
	docker tag blinkfile ${CONTAINER_REGISTRY}/blinkfile
	docker push ${CONTAINER_REGISTRY}/blinkfile

test:
	go test -coverprofile coverage.out ./...
	go tool cover -html=coverage.out -o=coverage.html