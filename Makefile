.PHONY: build run clean

build:
	go build -o ./bin/demo ./cmd/demo

run: build
	./bin/demo

clean:
	rm -rf ./bin ./output
