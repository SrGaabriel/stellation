set shell := ["bash", "-euo", "pipefail", "-c"]

default:
    @just --list

build:
    go build -o build/stellation .

test:
    go test ./...

check:
    test -z "$(gofmt -l .)" || { gofmt -l .; exit 1; }
    go vet ./...
    go test ./...

snapshot:
    goreleaser release --snapshot --clean
