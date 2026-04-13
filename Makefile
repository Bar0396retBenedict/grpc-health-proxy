.PHONY: build test lint run clean

BINARY := grpc-health-proxy
CMD     := ./cmd/grpc-health-proxy

build:
	go build -o bin/$(BINARY) $(CMD)

test:
	go test ./...

test-integration:
	go test -tags=integration ./cmd/grpc-health-proxy/...

lint:
	golangci-lint run ./...

run: build
	./bin/$(BINARY) \
		-http-addr 0.0.0.0:8080 \
		-grpc-addr localhost:50051 \
		-dial-timeout 5s \
		-check-interval 10s

clean:
	rm -rf bin/
