APP_NAME := ssmi
DIST_DIR := dist

clean:
	rm -rf $(DIST_DIR)

build:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(APP_NAME)

run: 
	go run ssmi.go
