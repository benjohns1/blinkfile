FROM golang:1.21-alpine
ARG ORIG_DIR=/orig_src
RUN apk add --update npm rsync
WORKDIR $ORIG_DIR
COPY app/web/package.json app/web/package-lock.json app/web/
RUN cd app/web && npm ci
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ENV COPY_DIR=/src
CMD ["sh", "-c", "mkdir -p ${COPY_DIR} && rsync -ar ./ ${COPY_DIR}"]
