
all: build install

build:
    go build -o ./bin/cli ./cmd/cli/main.go
install:
    go install ./cmd/cli/main.go
test:
    go test ./...
lint:
    golangci-lint run ./...
clean:
    rm -rf ./bin/cli