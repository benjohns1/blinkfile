---
title: Getting Started
weight: 200
---
## Docker Image
[On DockerHub](https://hub.docker.com/repository/docker/benjohns1/blinkfile)

```sh
docker run -p 8000:8000 -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=supersecretpassword benjohns1/blinkfile
```

### Persist data with a volume
```sh
docker volume create bf-data
docker run -p 8000:8000 -e ADMIN_USERNAME=admin -e ADMIN_PASSWORD=supersecretpassword -v bf-data:/data benjohns1/blinkfile
```

## Docker Compose
```yaml
services:
  blinkfile:
    image: benjohns1/blinkfile:latest
    volumes:
      - bf-data:/data
    environment:
      - ADMIN_USERNAME=bf_admin
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
    ports:
      - "8000:8000"
volumes:
  bf-data:
```
## Configuration
See [Environment Variables](./environment-variables) for a list of available configuration options.
