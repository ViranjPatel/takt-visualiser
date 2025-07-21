#!/bin/bash

# Start all services for development

echo "Starting Takt Visualiser development environment..."

# Check prerequisites
if ! command -v docker-compose &> /dev/null; then
    echo "docker-compose is required but not installed. Aborting."
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "Go is required but not installed. Aborting."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "npm is required but not installed. Aborting."
    exit 1
fi

# Start Docker services
docker-compose up -d postgres redis

echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Run database migrations
echo "Running database migrations..."
docker-compose exec -T postgres psql -U takt_user -d takt_visualiser < database/schema.sql
docker-compose exec -T postgres psql -U takt_user -d takt_visualiser < database/seed.sql

# Start backend
echo "Starting backend..."
cd backend
go mod download
DATABASE_URL="postgres://takt_user:takt_pass@localhost:5432/takt_visualiser?sslmode=disable" \
REDIS_URL="redis://localhost:6379" \
go run cmd/server/main.go &
BACKEND_PID=$!

# Start frontend
echo "Starting frontend..."
cd ../frontend
npm install
REACT_APP_API_URL="http://localhost:8080/api/v1" npm start &
FRONTEND_PID=$!

echo ""
echo "========================================"
echo "Development environment started!"
echo "Frontend: http://localhost:3000"
echo "Backend: http://localhost:8080"
echo "Press Ctrl+C to stop all services"
echo "========================================"
echo ""

# Cleanup function
cleanup() {
    echo "\nShutting down services..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    docker-compose down
    exit 0
}

# Set trap for cleanup
trap cleanup INT TERM

# Wait for processes
wait
