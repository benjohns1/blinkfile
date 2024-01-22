FROM blinkfile-src as build
ARG SRC_DIR=/src
WORKDIR $SRC_DIR
ARG CGO_ENABLED=0
ARG GOARCH=amd64
ARG GOOS=linux
RUN CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=${GOOS} go build -o /bin/binary ./cmd/main.go

FROM scratch
WORKDIR /
COPY --from=build /bin/binary .
EXPOSE 8000
CMD ["/binary"]