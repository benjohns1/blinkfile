run:
	go run cmd/main.go

test:
	go test -coverprofile coverage.out ./...
	go tool cover -html=coverage.out -o=coverage.html