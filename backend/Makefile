.PHONY: build run identity wallet payments marketplace provision settlement notify clean test

SERVICES := identity wallet payments marketplace provision settlement notify

build: $(SERVICES)

$(SERVICES):
	go build -o bin/$@ ./cmd/$@

run-%: build
	PORT_$*=8080 ./bin/$*

run-all:
	docker compose up --build

test:
	go vet ./...
	go test ./...

clean:
	rm -rf bin
