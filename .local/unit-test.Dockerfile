FROM golang:1.23.4-alpine
ENV GOCACHE=/cache/go-build
RUN mkdir -p /cache/go-build
ENV GOMODCACHE=/cache/go-mod
RUN mkdir -p /cache/go-mod
WORKDIR /src
CMD ["go", "test", "./..."]