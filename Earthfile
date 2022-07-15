FROM golang:1.18
WORKDIR /mixtool

deps:
    COPY go.mod go.sum ./
    RUN go mod download -x
    # Output these back in case go mod download changes them.
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum
    SAVE IMAGE --cache-hint

test:
    FROM +deps
    COPY . ./
    RUN make test

build:
    FROM +deps
    COPY . ./
    RUN make build
    SAVE ARTIFACT _output/linux/amd64/mixtool AS LOCAL build/mixtool
    SAVE ARTIFACT _output/linux/amd64/mixtool mixtool
    SAVE IMAGE --cache-hint
