${GOOS}BIN_NAME=yap
.DEFAULT_GOAL := run

# Detect platform
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	GOOS := linux
else ifeq ($(UNAME_S),Darwin)
	GOOS := darwin
else ifeq ($(UNAME_S),Windows)
	GOOS := windows
endif

build:
	GOARCH=amd64 GOOS=${GOOS} go build -o ./target/${BIN_NAME}-${GOOS} main.go

run: build
	./target/${BIN_NAME}-${GOOS}

build_and_run: build run

clean:
	go clean
	rimraf ./target

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

dep:
	go mod download

vet:
	go vet