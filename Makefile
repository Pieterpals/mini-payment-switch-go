.PHONY: run build test lint clean infra-up infra-down

# Application
run:
	go run ./cmd/api/...

build:
	go build -o bin/payment-switch ./cmd/api/...

test:
	go test ./... -v -cover -count=1

lint:
	go vet ./...

clean:
	rm -rf bin/

# Infrastructure
infra-up:
	docker-compose up -d

infra-down:
	docker-compose down

# Database Migration (manual)
migrate:
	@echo "Run the SQL files in migrations/ against your PostgreSQL database"
	@echo "Example: psql -h localhost -U admin -d payment_db -f migrations/001_create_transactions.sql"

# Dependencies
deps:
	go mod tidy
	go mod verify
