# Blinkfile
Send files quickly and securely.

![Blinkfile](docs/images/logo-light-bg-240x100.png)

## Features
- Upload a file and generate sharable links
- Single admin user authentication
- File expiration time can be set by duration or date
- Password-protect file access

## Self-host
Easy [docker compose setup](examples/docker-compose.yml) for your home lab

## Official image
[On DockerHub](https://hub.docker.com/repository/docker/benjohns1/blinkfile)

```
docker pull benjohns1/blinkfile
```

## Run locally
### With Docker
Prerequisites: Docker, Make

#### Build & run
```
make
```
#### Run unit tests & build
```
make build
```
#### Run unit tests, build, & run acceptance tests
```
make ci
```

### On host machine
Prerequisites: Go, NPM, Make

#### Install
```
make install
```
#### Unit test with coverage
```
make test
```
#### Run acceptance tests in a headless browser
```
make test-acceptance
```
#### Open Cypress to run acceptance tests
```
make test-acceptance-open
```
#### Run
Required environment variables:
- ADMIN_USERNAME
- ADMIN_PASSWORD
```
go run ./...
```