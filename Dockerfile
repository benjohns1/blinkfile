FROM golangci/golangci-lint:v1.55.2 AS linter
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN golangci-lint run

FROM golang:1.21-alpine AS tester
WORKDIR /src
COPY --from=linter /src /src
RUN go test ./...

FROM tester AS builder
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /app ./cmd/main.go
RUN chmod +x /app

FROM scratch AS release
WORKDIR /
COPY --from=builder /app /app
EXPOSE 8000
CMD ["/app"]