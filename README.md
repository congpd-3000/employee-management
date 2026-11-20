# Employee Management System

A pure Go REST API for managing employees and departments. The project uses only the Go standard library and follows the course requirement of a Handler → Service → Repository architecture.

## Technology stack

- `net/http` for the HTTP server and routing
- `encoding/json` for request, response, persistence, and export data
- `encoding/csv` for CSV export
- `context` for request propagation and timeouts
- `sync` and goroutines for storage locking and concurrent exports
- `testing` and `httptest` for automated tests

The application uses file-based storage (Option B). It does not require MySQL, Docker, third-party packages, environment variables, authentication, or optional middleware.

## Architecture

```text
Handler → Service → Repository
```

- **Handler layer** ([internal/handlers](internal/handlers)) parses HTTP requests and writes JSON responses.
- **Service layer** ([services](services)) contains validation and business logic.
- **Repository layer** ([repositories](repositories)) implements file storage behind repository interfaces.
- **Models** ([models](models)) define `Employee`, `EmployeeUpdate`, and `Department`.
- **Utilities** ([utils](utils)) provide the shared JSON response format.

## Project layout

```text
cmd/app/main.go              Application entry point
internal/handlers/           HTTP routes and handlers
models/                      Employee and Department models
repositories/                Repository interfaces and JSON file storage
services/                    Business logic and concurrent export
utils/                       JSON response helpers
data/                        Runtime JSON storage
exports/                     Export destination directory
```

## Requirements

- Go 1.25 or later

No database, Docker container, `.env` file, or dependency installation is required.

## Running the application

From the repository root:

```bash
go run ./cmd/app
```

The server listens at:

```text
http://localhost:3000
```

The application loads data from `data/storage.json` when it starts and saves the file after create, update, and delete operations. The file is created automatically when the first record is saved and is excluded from Git.

The root path `/` is not an application route, so opening it returns `404 page not found`.

## API endpoints

### Employees

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/employees` | Create an employee |
| `GET` | `/employees?limit=10&offset=0&departmentId=1` | List employees with pagination and optional department filtering |
| `GET` | `/employees/{id}` | Get an employee by ID |
| `PUT` | `/employees/{id}` | Update one or more employee fields |
| `DELETE` | `/employees/{id}` | Permanently delete an employee |
| `GET` | `/employees/search?keyword=engineer` | Search by name or position |

### Departments

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/departments` | Create a department |

No endpoint requires authentication.

## Request examples

Create a department:

```bash
curl -X POST http://localhost:3000/departments \
  -H "Content-Type: application/json" \
  -d '{"name":"Engineering"}'
```

Create an employee:

```bash
curl -X POST http://localhost:3000/employees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nguyen Van A",
    "age": 28,
    "position": "Backend Engineer",
    "departmentId": 1,
    "salary": 2000
  }'
```

List employees:

```bash
curl "http://localhost:3000/employees?limit=10&offset=0&departmentId=1"
```

Partially update an employee:

```bash
curl -X PUT http://localhost:3000/employees/1 \
  -H "Content-Type: application/json" \
  -d '{"salary":2500}'
```

Search employees:

```bash
curl "http://localhost:3000/employees/search?keyword=engineer"
```

Delete an employee:

```bash
curl -X DELETE http://localhost:3000/employees/1
```

## Response formats

The paginated employee list has this shape:

```json
{
  "totalCount": 1,
  "employees": [
    {
      "id": 1,
      "name": "Nguyen Van A",
      "age": 28,
      "position": "Backend Engineer",
      "departmentId": 1,
      "salary": 2000
    }
  ]
}
```

Errors use a consistent JSON shape:

```json
{
  "error": "Employee not found",
  "code": 404
}
```
