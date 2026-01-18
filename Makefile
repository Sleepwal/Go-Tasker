.PHONY: build test fmt lint clean

build:
	go build -o bin/tasker ./cmd/tasker

test:
	go test -race ./...

fmt:
	gofmt -w .

lint:
	golangci-lint run

clean:
	rm -rf bin/
	rm -f coverage.out
