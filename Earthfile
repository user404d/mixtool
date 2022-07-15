FROM golang:1.17
WORKDIR /mixtool

deps:
    COPY go.mod go.sum ./
    RUN --mount=type=cache,target=/go/pkg/mod go mod download
    # Output these back in case go mod download changes them.
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum
    SAVE IMAGE --cache-hint

test:
    FROM +deps
    COPY Makefile *.go ./
    RUN make test

build:
    FROM +deps
    COPY Makefile VERSION go.mod go.sum *.go ./
    RUN --mount=type=cache,target=/mixtool/.cache/go-build make build
    SAVE ARTIFACT _output/linux/amd64/mixtool AS LOCAL build/mixtool
    SAVE ARTIFACT _output/linux/amd64/mixtool mixtool
    SAVE IMAGE --cache-hint
