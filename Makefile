BINARY_NAME=KabardaApp

build:
	@go mod vendor
	@echo "Building Kabarda..."
	@go build -o tmp/${BINARY_NAME} .
	@echo "Kabarda built!"

run: build
	@echo "Starting Kabarda..."
	@./tmp/${BINARY_NAME} &
	@echo "Kabarda started!"

clean:
	@echo "Cleaning..."
	@go clean
	@rm tmp/${BINARY_NAME}
	@echo "Cleaned!"

# Docker
start_compose:
	sudo docker-compose up -d

stop_compose:
	sudo docker-compose down
test:
	@echo "Testing..."
	@go test ./...
	@echo "Done!"

start: run

stop:
	@echo "Stopping Kabarda..."
	@-pkill -SIGTERM -f "./tmp/${BINARY_NAME}"
	@echo "Stopped Kabarda!"

restart: stop start