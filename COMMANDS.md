# Project Commands Guide

This project uses `make` to simplify common development tasks. All the commands below can be executed from the root directory of the project using your terminal.

## Running the Application

| Command | Description |
|---|---|
| `make run` | Starts the Go application locally (runs `main.go` inside `./cmd/api`). |

## Dependency Management

| Command | Description |
|---|---|
| `make tidy` | Cleans up and updates the `go.mod` and `go.sum` files, removing unused dependencies. |
| `make download` | Downloads all the Go modules required by the project. |

## Swagger

| Command | Description |
|---|---|
| `make swagger` | Generates Swagger documentation for the API. |

## Testing

The project has various test commands to run unit tests and integration tests with coverage:

| Command | Description |
|---|---|
| `make test` | Runs all tests in the project. |
| `make test-all` | Runs all unit and integration tests and calculates test coverage across handlers, services, and repositories. |
| `make test-unit-cover` | Runs all unit tests and checks their code coverage. |
| `make test-unit-service` | Runs only the unit tests for the **services** layer with coverage tracking. |
| `make test-unit-handler` | Runs only the unit tests for the **handlers** layer with coverage tracking. |
| `make test-integration` | Runs the integration tests with coverage tracking. |

## Docker & Container Management

Use these commands to manage the application's infrastructure using Docker Compose:

| Command | Description |
|---|---|
| `make docker-up` | Starts all the Docker containers (like database, Redis, etc.) in the background (detached mode). |
| `make docker-down` | Stops and removes the Docker containers, networks, and volumes created by `make docker-up`. |
| `make docker-logs` | Follows and displays the live logs output from all running Docker containers. |

---

