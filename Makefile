# Variables
APP_NAME := app
SRC := main.go

#? build: Build the project and create app binary
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(SRC)

#? run: Run the application
run: build
	@echo "Running $(APP_NAME)..."
	./$(APP_NAME)

#? test: Run tests with verbose output
test:
	go test ./... -v

#? cover: Generate the test coverage
cover:
	go test -cover ./...

#? lint: Run golangci-lint
lint:
	golangci-lint run -v

#? fmt: Format the code
fmt:
	go fmt ./...

#? clean: Clean binaries
clean:
	@rm -f $(APP_NAME)

#? all: Run all targets
all: fmt lint test run

#? help: List all available make targets with their descriptions
help: Makefile
	@$(call print, "Listing commands:")
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'