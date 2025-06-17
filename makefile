APP_NAME := ssmi
DIST_DIR := dist
VERSION := 1.0.0

default:
	@echo "Usage: make [target]"
	@echo "Available targets:"
	@echo "  build   - Build the application"
	@echo "  run     - Run the application"
	@echo "  clean   - Clean up build artifacts"
	@echo "  all     - Clean, build, and run the application"

all: clean build run

clean:
	rm -rf $(DIST_DIR)

build:
	mkdir -p $(DIST_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(APP_NAME)
	

run: 
	go run ssmi.go $(ARGS)
