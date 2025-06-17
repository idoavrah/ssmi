APP_NAME := ssmi
DIST_DIR := dist

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
	go build -o $(DIST_DIR)/$(APP_NAME)

run: 
	go run ssmi.go
