.PHONY: build test clean

build:
	go build -o bin/rpi ./cmd/rpi

test:
	go test ./...

clean:
	rm -f bin/rpi
