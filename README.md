# Takt Visualiser

A lightweight, high-performance Takt planning visualiser built for global scale.

## Architecture

- **Frontend**: React + DayPilot Scheduler
- **Backend**: Go REST API
- **Database**: PostgreSQL
- **Cache**: Redis
- **CDN**: Cloudflare

## Key Features

- Handles 100k+ tasks with smooth performance
- Sub-50ms API response times globally
- Real-time updates via WebSocket
- Virtual scrolling for massive datasets
- Delta-based syncing for minimal bandwidth

## Project Structure

```
/
├── backend/          # Go API server
├── frontend/         # React application
├── database/         # SQL schemas and migrations
├── docker/           # Container configurations
└── scripts/          # Build and deployment scripts
```

## Quick Start

```bash
# Backend
cd backend
go mod download
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm start
```

## Performance Targets

- API Response: < 50ms (p99)
- Frontend Render: < 16ms per frame
- Memory Usage: < 500MB for 100k tasks
- Concurrent Users: 100+
