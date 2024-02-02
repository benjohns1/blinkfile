FROM golang:1.21-alpine as source
RUN apk add --update npm rsync
ARG GOCACHE=/cache/go-build
ENV GOCACHE=${GOCACHE}
RUN mkdir -p ${GOCACHE}
ARG GOMODCACHE=/cache/go-mod
ENV GOMODCACHE=${GOMODCACHE}
RUN mkdir -p ${GOMODCACHE}
ARG SRC_DIR=/src
WORKDIR $SRC_DIR
COPY app/web/package.json app/web/package-lock.json app/web/
RUN cd app/web && npm ci
COPY go.* ./
RUN go mod download
COPY . ./
ENV OUT_DIR=/out_src
ENV OUT_GO_CACHE=/out_cache
CMD ["sh", "-c", "mkdir -p ${OUT_DIR} && rsync -ar ./ ${OUT_DIR} && rsync -ar /cache/ ${OUT_GO_CACHE}"]

FROM source as build
ARG SRC_DIR=/src
WORKDIR $SRC_DIR
ARG CGO_ENABLED=0
ARG GOARCH=amd64
ARG GOOS=linux
RUN CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=${GOOS} go build -o /bin/binary ./cmd/main.go

FROM scratch
WORKDIR /
COPY --from=build /bin/binary .
EXPOSE 8020
CMD ["/binary"]