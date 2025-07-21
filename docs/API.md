# Takt Visualiser API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Endpoints

### Health Check
```http
GET /health
```

Returns server status.

**Response:**
```json
{
  "status": "healthy",
  "time": 1234567890
}
```

### Tasks

#### Get Tasks
```http
GET /tasks?project_id=1&zone_ids=5,6&date_from=2025-08-01&date_to=2025-08-31
```

**Query Parameters:**
- `project_id` (required): Project ID
- `zone_ids` (optional): Comma-separated zone IDs
- `date_from` (optional): Start date filter (YYYY-MM-DD)
- `date_to` (optional): End date filter (YYYY-MM-DD)

**Response:** Array of tasks

#### Get Single Task
```http
GET /tasks/{id}
```

#### Update Task
```http
PATCH /tasks/{id}
```

**Request Body:**
```json
{
  "name": "Updated task name",
  "start_date": "2025-08-15",
  "duration": 5,
  "status": "in-progress"
}
```

#### Bulk Update Tasks
```http
PATCH /tasks/bulk
```

**Request Body:**
```json
[
  {
    "id": 1,
    "status": "completed"
  },
  {
    "id": 2,
    "start_date": "2025-08-20"
  }
]
```

### Zones

#### Get Zone Tree
```http
GET /zones/{projectId}/tree
```

Returns hierarchical zone structure.

#### Create Zone
```http
POST /zones
```

**Request Body:**
```json
{
  "project_id": 1,
  "parent_id": 2,
  "name": "Zone 1C",
  "level": 2
}
```

### WebSocket

```
ws://localhost:8080/api/v1/ws
```

Connects to real-time update stream.

**Message Format:**
```json
{
  "type": "task_update",
  "task_id": "123"
}
```

## Performance Guarantees

- All GET endpoints: < 50ms response time (p99)
- Bulk operations: < 200ms for up to 1000 items
- WebSocket latency: < 10ms for broadcasts
- Supports 100+ concurrent connections
