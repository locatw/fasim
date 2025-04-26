.PHONY: dev build clean test

# Start development environment
dev: build
	docker compose up

# Build frontend and backend
build:
	cd frontend && npm run build
	cd backend && go build -o bin/server cmd/server/main.go

# Build frontend
build-frontend:
	cd frontend && npm run build

# Build backend
build-backend:
	cd backend && go build -o bin/server cmd/server/main.go

# Cleanup
clean:
	rm -rf frontend/build
	rm -rf backend/bin

# Run tests
test: build
	cd frontend && npm test
	cd backend && go test ./...

# Run frontend tests
test-frontend: build-frontend
	cd frontend && npm test

# Run backend tests
test-backend: build-backend
	cd backend && go test ./...

# Install frontend dependencies
install-frontend:
	cd frontend && npm install

# Install backend dependencies
install-backend:
	cd backend && go mod tidy

# Install dependencies
install: install-frontend install-backend
