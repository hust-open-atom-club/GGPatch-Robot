.PHONY: build test clean

all: build

build:
	go build -o ggpatch-robot ./cmd/ggpatch-robot

test:
	go test ./...

clean:
	rm -f ggpatch-robot
