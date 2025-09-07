BINARY_NAME=bai
MAIN_PACKAGE=./main.go

.PHONY: build
build:
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

.PHONY: install
install:
	go install $(MAIN_PACKAGE)

.PHONY: clean
clean:
	go clean
	rm $(BINARY_NAME)

.PHONY: test
test:
	go test -v ./...

.PHONY: run
run:
	go run $(MAIN_PACKAGE)

.PHONY: deps
deps:
	go mod download
	go mod tidy
