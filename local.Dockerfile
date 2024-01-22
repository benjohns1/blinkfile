FROM golang:1.21-alpine AS setup
RUN apk add --update npm
WORKDIR /src
COPY app/web/package.json ./app/web/package-lock.json /src/app/web/
RUN cd app/web && npm ci
COPY go.mod go.sum /src/
RUN go mod download
COPY . ./

FROM golangci/golangci-lint:v1.55.2 AS linter
WORKDIR /src
COPY --from=setup /src /src
RUN golangci-lint run

FROM setup AS unit-tester
COPY --from=setup /src /src
RUN go test ./...

FROM setup AS builder
COPY --from=setup /src /src
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /app ./cmd/main.go
RUN chmod +x /app

FROM scratch AS release
WORKDIR /
COPY --from=builder /app /app
EXPOSE 8000
CMD ["/app"]