.PHONY: run docker-up docker-down

run:
	@echo "Starting application..."
	@go run ./backend/cmd/server/main.go

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

