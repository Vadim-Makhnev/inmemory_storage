.PHONY: test build run docker-build docker-run docker-logs clean

test:
	go test -v ./...

run: build
	@./bin/redisclone

build:
	@go build -o bin/redisclone ./cmd/server

docker-run: docker-build
	docker run --rm -d -p 4000:4000 --name inmemory inmemory

docker-build:
	docker build -t inmemory .

docker-stop:
	docker stop inmemory

docker-logs:
	docker logs -f inmemory

clean:
	rm -f bin/redisclone
