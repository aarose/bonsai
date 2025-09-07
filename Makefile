BINARY_NAME=bai
MAIN_PACKAGE=./main.go

.PHONY: build
build:
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

.PHONY: install
install:
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)
	cp $(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME)

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

.PHONY: release
release: test
	@echo "Building release binaries..."
	mkdir -p dist
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o dist/bai-linux-amd64 $(MAIN_PACKAGE)
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -o dist/bai-darwin-amd64 $(MAIN_PACKAGE)
	# macOS ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 go build -o dist/bai-darwin-arm64 $(MAIN_PACKAGE)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -o dist/bai-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Release binaries created in dist/"
