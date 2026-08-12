# IncidentHub

A full-stack incident management platform for tracking, managing, and resolving incidents.

## Features

- **Authentication**: Email/password registration and login with JWT tokens
- **Incidents CRUD**: Create, view, update, and delete incidents
- **Incident Management**: Assign to users, change severity (LOW/MEDIUM/HIGH/CRITICAL) and status (OPEN/INVESTIGATING/MITIGATED/RESOLVED)
- **Comments**: Add comments to incidents for collaboration
- **Dashboard**: Real-time statistics (total, open, investigating, critical, resolved)
- **Search & Filter**: Server-side search, filter by status/severity/assignee, pagination
- **Responsive UI**: Clean, modern dashboard built with Next.js

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend   │────▶│   Backend   │────▶│  PostgreSQL  │
│  (Next.js)   │◀────│  (Go/Gin)   │◀────│   Database   │
└─────────────┘     └─────────────┘     └─────────────┘
```

- **Backend**: Go REST API using Gin framework + sqlx
- **Frontend**: Next.js 15 with TypeScript
- **Database**: PostgreSQL with migrations
- **Containerization**: Docker Compose for local deployment

## Quick Start

### Prerequisites

- Docker and Docker Compose

### Running with Docker Compose

```bash
# Copy environment variables
cp .env.example .env

# Build and start all services
docker compose up --build

# Services available at:
# - Frontend: http://localhost:3000
# - Backend API: http://localhost:8080
# - PostgreSQL: localhost:5432
```

### Running Locally (Development)

#### Backend

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Login and get JWT token |

### Incidents
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/incidents` | List incidents (with search/filter/pagination) |
| POST | `/api/incidents` | Create a new incident |
| GET | `/api/incidents/:id` | Get incident details |
| PUT | `/api/incidents/:id` | Update an incident |
| DELETE | `/api/incidents/:id` | Delete an incident |

### Comments
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/incidents/:id/comments` | List comments for an incident |
| POST | `/api/incidents/:id/comments` | Add a comment to an incident |

### Dashboard
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/dashboard/stats` | Get dashboard statistics |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@db:5432/incidenthub?sslmode=disable` |
| `JWT_SECRET` | Secret key for JWT signing | (required) |
| `PORT` | Backend server port | `8080` |
| `FRONTEND_URL` | Frontend URL for CORS | `http://localhost:3000` |

## Testing

### Backend Tests

```bash
cd backend
go test ./...
```

### Frontend Tests

```bash
cd frontend
npm test
```

## Project Structure

```
incidenthub/
├── backend/
│   ├── cmd/server/        # Application entry point
│   ├── internal/
│   │   ├── auth/          # JWT authentication
│   │   ├── handlers/      # HTTP handlers
│   │   ├── middleware/    # Gin middleware
│   │   ├── migrations/    # Database migrations
│   │   └── models/        # Data models
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/app/           # Next.js App Router
│   ├── Dockerfile
│   └── package.json
├── docker-compose.yml
├── .env.example
└── README.md
```

## License

MIT