.PHONY: all run test
all: run

run:
	go run ./cmd/lanfile

test:
	go test ./...