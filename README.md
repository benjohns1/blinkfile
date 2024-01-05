# BlinkFile
Send files quickly and securely.

## Features
- Upload a file and generate sharable links
- Single admin user authentication
- File expiration time can be set by duration or date
- Password-protect file access

## Self-host
Easy [docker compose setup](examples/docker-compose.yml) for your home lab

## Running locally
### With Docker
Prerequisites: Docker, Make
#### Test, build & run
```
make
```
#### Test & build
```
make build
```
### On host
Prerequisites: Go, NPM, Make

Required environment variables:
- ADMIN_USERNAME
- ADMIN_PASSWORD
#### Test with unit coverage
```
make test
```
#### Build
```
go build -o /app ./cmd/main.go
```