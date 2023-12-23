build:
	docker build --tag one-way-file-send .

run: build
	docker volume create one-way-file-send-data
	docker run --name one-way-file-send -p 8000:8000 -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=1234123412341234 -e DATA_DIR=/data -v one-way-file-send-data:/data --rm one-way-file-send

test:
	go test -coverprofile coverage.out ./...
	go tool cover -html=coverage.out -o=coverage.html