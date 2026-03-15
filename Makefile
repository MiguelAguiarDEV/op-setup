.PHONY: build test test-cover lint clean run

BINARY := op-setup
MODULE := github.com/MiguelAguiarDEV/op-setup
VERSION ?= dev

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/op-setup

run: build
	./$(BINARY)

test:
	go test ./... -count=1 -race

test-cover:
	go test ./... -count=1 -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	go vet ./...

clean:
	rm -f $(BINARY) coverage.out coverage.html

update-golden:
	UPDATE_GOLDEN=1 go test ./... -count=1 -run Golden
