
all: build install

build:
    go build -o ./bin/envseal-cli ./cmd/envseal-cli/main.go
install:
    go install ./cmd/envseal-cli/main.go
test:
    go test ./...
lint:
    golangci-lint run ./...
clean:
    rm -rf ./bin/envseal-cli