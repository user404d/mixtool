FROM golang:1.17

ARG WORKDIR="/mixtool"
WORKDIR $WORKDIR

ARG GOMODCACHE="/go/pkg/mod"
ENV GOMODCACHE=$GOMODCACHE
ARG GOCACHE="$WORKDIR/.cache/go-build"
ENV GOCACHE=$GOCACHE

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    # Output these back in case go mod download changes them.
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum
    SAVE IMAGE --cache-hint

test:
    FROM +deps
    COPY . .
    RUN make test

build:
    FROM +deps
    COPY go.mod go.sum Makefile VERSION .
    COPY --dir cmd pkg .
    RUN --mount=type=cache,target=$GOCACHE make build
    SAVE ARTIFACT _output/linux/amd64/mixtool mixtool AS LOCAL build/mixtool
    SAVE IMAGE --cache-hint
