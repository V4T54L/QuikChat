.PHONY: up down logs psql migrate-up run build

# Docker Compose commands
up:
	@echo "Starting Docker containers..."
	docker-compose up -d

down:
	@echo "Stopping Docker containers..."
	docker-compose down

logs:
	@echo "Tailing logs..."
	docker-compose logs -f

# Database commands
psql:
	@echo "Connecting to PostgreSQL container..."
	docker-compose exec postgres psql -U user -d chatdb

migrate-up:
	@echo "Applying database migrations..."
	docker-compose exec -T postgres psql -U user -d chatdb < backend/migrations/000001_create_users_table.up.sql
	docker-compose exec -T postgres psql -U user -d chatdb < backend/migrations/000002_create_sessions_table.up.sql

# Go application commands
run:
	@echo "Running the Go application..."
	go run ./backend/cmd/server/main.go

build:
	@echo "Building the Go application..."
	go build -o ./bin/server ./backend/cmd/server/main.go

