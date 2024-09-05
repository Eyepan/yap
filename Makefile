BIN_NAME=yap
.DEFAULT_GOAL := run

build:
	GOARCH=amd64 GOOS=darwin go build -o ./target/${BIN_NAME}-darwin main.go
	GOARCH=amd64 GOOS=windows go build -o ./target/${BIN_NAME}-windows main.go
	GOARCH=amd64 GOOS=linux go build -o ./target/${BIN_NAME}-linux main.go

run: build
	./target/${BIN_NAME}-linux

build_and_run: build run

clean:
	go clean
	rimraf ./target

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

dep:
	go mod downlaoad

vet:
	go vet
