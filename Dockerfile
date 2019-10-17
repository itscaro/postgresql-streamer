# Build stage
FROM golang:1.13 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ARG GIT_COMMIT
ARG VERSION
ARG OS

COPY . .
RUN go test -tags=test -v -covermode=count -coverprofile="cover.out" ./... && \
    go tool cover -func="cover.out"
RUN [ "$OS" = "Linux" -o -z "$OS" ] \
    && { echo "✅ Build for Linux"; \
    CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
        -X jgithub.com/allocine/postgresql-streamer-go/utils.GitCommit=${GIT_COMMIT} \
        -X jgithub.com/allocine/postgresql-streamer-go/utils.Version=${VERSION}" \
        -a -installsuffix cgo -o build/postgresql-streamer-amd64 || exit 1; } \
    || echo "⛔ Skip build for Linux"; \
#
    [ "$OS" = "Darwin" -o -z "$OS" ] \
    && { echo "✅ Build for Darwin"; \
    CGO_ENABLED=0 GOOS=darwin go build --ldflags "-s -w \
        -X jgithub.com/allocine/postgresql-streamer-go/utils.GitCommit=${GIT_COMMIT} \
        -X jgithub.com/allocine/postgresql-streamer-go/utils.Version=${VERSION}" \
        -a -installsuffix cgo -o build/postgresql-streamer-darwin || exit 1; } \
    || echo "⛔ Skip build for Darwin"; \
#
    [ "$OS" = "Windows_NT" -o -z "$OS" ] \
    && { echo "✅ Build for Windows"; \
    CGO_ENABLED=0 GOOS=windows go build --ldflags "-s -w \
        -X jgithub.com/allocine/postgresql-streamer-go/utils.GitCommit=${GIT_COMMIT} \
        -X jgithub.com/allocine/postgresql-streamer-go/utils.Version=${VERSION}" \
        -a -installsuffix cgo -o build/postgresql-streamer.exe || exit 1; } \
    || echo "⛔ Skip build for Windows"

# Release stage
FROM alpine:3.9

RUN apk --no-cache add ca-certificates git make postgresql-client

RUN mkdir /app

WORKDIR /app

COPY --from=builder /app/build/. /app/
COPY --from=builder /app/Makefile /app/

ENTRYPOINT ["/app/postgresql-streamer-amd64"]
