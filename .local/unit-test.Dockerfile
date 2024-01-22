FROM golang:1.21-alpine
ENV GOCACHE=/cache/go-build
RUN mkdir -p /cache/go-build
ENV GOMODCACHE=/cache/go-mod
RUN mkdir -p /cache/go-mod
ARG SRC_DIR=/src
WORKDIR $SRC_DIR
CMD ["go", "test", "./..."]
