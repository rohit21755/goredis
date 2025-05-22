run: build
	@./bin/goredis --listenAddr :8000
build:
	@go build -o bin/goredis .
