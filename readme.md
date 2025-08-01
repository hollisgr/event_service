# Event Service

## Description
Event Service is a Go-based application designed for event processing. It includes functionality for:
- Database interaction
- Data validation
- Event handler registration
- Configuration management
- Scheduling tasks

## Technology Stack
- **Go** — Programming language.
- **Gin** — Web framework for building REST APIs.
- **PostgreSQL** — Database.
- **Validator** — Library for data validation.
- **Logrus** — Library for logging.

## Installation and Setup

### Requirements
- Docker installed and running.
- Docker Compose installed.

### Steps to Install
1. **Clone the repository**:
   git clone <repository-url>
   cd <repository-name>

2. **Build and run the application using Docker Compose**:
   docker-compose up --build

3. **Access the application**:
   The service will be available at `http://localhost:<port>` (replace `<port>` with the port specified in the `docker-compose.yml` file).

4. **Stop the application**:
   docker-compose down

## Project Structure
- `internal/cfg` — Configuration management module.
- `internal/db` — Database interaction module.
- `internal/listener/handler_service` — Event handler module.
- `internal/scheduler` — Task scheduling module.
- `internal/pipeline` — Pipeline processing module.
- `pkg/logger` — Logging module.
- `pkg/postgres` — PostgreSQL client module.

## Features
- Connects to PostgreSQL database.
- Validates incoming data using `validator`.
- Registers event handlers.
- Logs errors and events using `Logrus`.
- Schedules and manages tasks using `scheduler`.